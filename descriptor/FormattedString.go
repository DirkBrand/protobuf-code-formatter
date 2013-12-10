package descriptor

import (
	"bytes"
	proto "code.google.com/p/gogoprotobuf/proto"
	"encoding/binary"
	fmt "fmt"
	sort "sort"
	strings "strings"
)

const (
	INDENT = "  "
)

const (
	VARINT_STRING  = 504026
	VARINT_INTBOOL = 504024
)

var importsList []string
var allFiles []*FileDescriptorProto
var currentFile FileDescriptorProto

// Handles the set of Files
func (this *FileDescriptorSet) FormattedGoString(fileToFormat string) string {
	// Loop through all the FileDescriptorProto
	allFiles = this.File
	for _, tmpFile := range this.File {
		if tmpFile.GetName() == fileToFormat {
			return tmpFile.FormattedGoString(0)
		}
	}
	return ""
}

// Handles entire file
func (this *FileDescriptorProto) FormattedGoString(depth int) string {
	if this == nil {
		return "nil"
	}
	currentFile = *this

	var s string

	// the package
	if len(this.GetPackage()) > 0 {
		s += `package ` + this.GetPackage() + "\n"
	}

	// For each import
	if len(this.GetDependency()) > 0 && len(this.GetPackage()) > 0 {
		s += "\n"
	}
	if len(this.GetDependency()) > 0 {
		importsList = make([]string, len(this.GetDependency()))
		i := 0

		sort.Strings(this.GetDependency())
		for _, imp := range this.GetDependency() {
			importsList[i] = strings.Split(imp, "/")[len(strings.Split(imp, "/"))-1]
			i += 1

			s += `import "` + imp + `";` + "\n"
		}
	}

	// For each extend
	extendGroups := make(map[string]string)
	for _, ext := range this.GetExtension() {
		extendGroups[ext.GetExtendee()] = extendGroups[ext.GetExtendee()] + ext.FormattedGoString(depth+1) + ";\n"

	}
	if len(extendGroups) > 0 {
		s += "\n"
	}
	for i := range extendGroups {
		group := extendGroups[i]
		s += getIndentation(depth) + `extend ` + strings.Replace(i, ".", "", 1) + " {\n"
		s += group
		s += getIndentation(depth) + "}\n"
	}

	// Field Options
	options := this.GetOptions()
	if options != nil && len(options.ExtensionMap()) > 0 {
		s += "\n"

		s += getFormattedOptionsFromExtensionMap(options.ExtensionMap(), -1, false)
	}

	// Enums
	if len(this.GetEnumType()) > 0 {
		s += "\n"
	}
	for _, enum := range this.GetEnumType() {
		s += enum.FormattedGoString(depth)
		s += "\n"
	}

	// Messages
	if len(this.GetMessageType()) > 0 {
		s += "\n"
	}
	for _, message := range this.GetMessageType() {
		s += message.FormattedGoString(depth, false)
		s += "\n"
	}

	for _, service := range this.GetService() {
		s += service.FormattedGoString(depth)
		s += "\n"
	}

	// SourceCodeInfo
	source := this.GetSourceCodeInfo()
	for _, sourceLoc := range source.GetLocation() {
		s += sourceLoc.String()
		// TODO comments
	}

	return s
}

// Handles Messages
func (this *DescriptorProto) FormattedGoString(depth int, isGroup bool) string {
	if this == nil {
		return "nil"
	}
	var s string
	contentCount := 0

	// For groups logic
	nestedMessages := this.GetNestedType()

	// Message Header
	if isGroup {
		s += getIndentation(depth) + `group ` + fmt.Sprintf("%v", this.GetName()) + ` {`
	} else {
		s += getIndentation(depth) + `message ` + fmt.Sprintf("%v", this.GetName()) + ` {`
	}

	// Extension Range
	if len(this.GetExtensionRange()) > 0 {
		s += "\n"
		contentCount += 1
	}
	for extensionIndex, ext := range this.GetExtensionRange() {
		s += ext.FormattedGoString(depth + 1)
		if extensionIndex < len(this.GetExtension())-1 {
			s += "\n"
		}
	}

	// For each extend
	if len(this.GetExtension()) > 0 {
		s += "\n"
		contentCount += 1
	}
	extendGroups := make(map[string]string)
	for _, ext := range this.GetExtension() {
		if depth == 0 {
			extendGroups[ext.GetExtendee()] = extendGroups[ext.GetExtendee()] + getIndentation(depth+1) + ext.FormattedGoString(depth+1) + ";\n"
		} else {
			extendGroups[ext.GetExtendee()] = extendGroups[ext.GetExtendee()] + getIndentation(depth) + ext.FormattedGoString(depth+1) + ";\n"
		}

	}
	for i := range extendGroups {
		group := extendGroups[i]
		s += getIndentation(depth+1) + `extend ` + getLastWordFromPath(i, ".") + " {\n"
		s += group
		s += getIndentation(depth+1) + "}\n"
	}

	// Options
	mesOptions := this.GetOptions()
	if mesOptions != nil && len(mesOptions.ExtensionMap()) > 0 {
		s += "\n"
		contentCount += 1

		s += getFormattedOptionsFromExtensionMap(mesOptions.ExtensionMap(), depth, false)
	}

	// Fields
	if len(this.GetField()) > 0 {
		s += "\n"
		contentCount += 1
	}
	for _, field := range this.GetField() {
		if field.GetType() == FieldDescriptorProto_TYPE_GROUP {
			for i := 0; i < len(nestedMessages); i += 1 {
				nestedMes := nestedMessages[i]
				// Found group
				if strings.ToLower(nestedMes.GetName()) == field.GetName() {
					s += "\n"
					tempStr := nestedMes.FormattedGoString(depth+1, true)
					tempStr = strings.Replace(tempStr, "group", fieldDescriptorProtoLabel_StringValue(field.GetLabel())+" group", 1)
					tempStr = strings.Replace(tempStr, nestedMes.GetName(), nestedMes.GetName()+" = "+fmt.Sprintf("%v", field.GetNumber()), 1)
					s += tempStr
					nestedMessages = append(nestedMessages[:i], nestedMessages[i+1:]...)
				}
			}
		} else {
			s += field.FormattedGoString(depth + 1)
			s += ";\n"
		}
	}

	// Enums
	if len(this.GetEnumType()) > 0 {
		s += "\n"
		contentCount += 1
	}
	for _, enum := range this.GetEnumType() {
		s += enum.FormattedGoString(depth + 1)
	}

	// Nested Messages
	if len(nestedMessages) > 0 {
		s += "\n"
		contentCount += 1
	}
	for _, nestedMessage := range nestedMessages {
		s += nestedMessage.FormattedGoString(depth+1, false)
	}

	if contentCount > 0 {
		s += getIndentation(depth)
	}
	s += "}\n"

	return s
}

// Handles Fields
func (this *FieldDescriptorProto) FormattedGoString(depth int) string {
	if this == nil {
		return "nil"
	}
	var s string
	s += getIndentation(depth) + fmt.Sprintf("%v", fieldDescriptorProtoLabel_StringValue(*this.Label))
	// If referencing a message
	if *this.Type == FieldDescriptorProto_TYPE_MESSAGE || *this.Type == FieldDescriptorProto_TYPE_ENUM {
		s += ` ` + strings.Replace(this.GetTypeName(), ".", "", 1)
	} else {
		s += ` ` + fmt.Sprintf("%v", fieldDescriptorProtoType_StringValue(*this.Type))
	}
	s += ` ` + fmt.Sprintf("%v", this.GetName()) + ` = ` + fmt.Sprintf("%v", this.GetNumber())

	// OPTIONS
	options := this.GetOptions()
	i := 0
	if options != nil {
		if options.GetPacked() || options.GetLazy() || options.GetDeprecated() || len(this.GetDefaultValue()) > 0 || len(options.ExtensionMap()) > 0 {
			s += ` [`

			if len(options.ExtensionMap()) > 0 {
				s += getFormattedOptionsFromExtensionMap(options.ExtensionMap(), -1, true)
				i += 1
			}

			if len(this.GetDefaultValue()) > 0 {
				if i >= 1 {
					s += ", "
				}
				s += `default = ` + this.GetDefaultValue()
			}

			if options.GetPacked() {
				if i >= 1 {
					s += ", "
				}
				s += `packed = true`
			}

			if options.GetLazy() {
				if i >= 1 {
					s += ", "
				}
				s += `lazy = true`
			}

			if options.GetDeprecated() {
				if i >= 1 {
					s += ", "
				}
				s += `deprecated = true`
			}

			s += `]`
		}
	}

	return s
}

// Handles Enums
func (this *EnumDescriptorProto) FormattedGoString(depth int) string {
	if this == nil {
		return "nil"
	}
	var s string
	s += getIndentation(depth) + `enum ` + fmt.Sprintf("%v", this.GetName()) + ` {`

	// Options
	options := this.GetOptions()
	if options != nil && len(options.ExtensionMap()) > 0 {
		s += "\n"

		s += getFormattedOptionsFromExtensionMap(options.ExtensionMap(), depth, false)
	}

	if len(this.GetValue()) > 0 {
		s += "\n"
	}
	for _, enumValue := range this.GetValue() {
		s += getIndentation(depth+1) + enumValue.GetName() + ` = ` + fmt.Sprintf("%v", enumValue.GetNumber())
		// OPTIONS
		valueOptions := enumValue.GetOptions()
		if valueOptions != nil {
			s += ` [` + getFormattedOptionsFromExtensionMap(valueOptions.ExtensionMap(), -1, true) + `]`
		}

		s += ";\n"
	}

	s += getIndentation(depth) + "}\n"

	return s
}

// Handles Extension Ranges
func (this *DescriptorProto_ExtensionRange) FormattedGoString(depth int) string {
	if this == nil {
		return "nil"
	}
	var s string
	s += getIndentation(depth) + `extensions ` + fmt.Sprintf("%v", this.GetStart()) + ` to `
	if this.GetEnd() >= 1<<29-1 {
		s += "max;\n"
	} else {
		s += fmt.Sprintf("%v", this.GetEnd()) + ";\n"
	}

	return s
}

// Handles Services
func (this *ServiceDescriptorProto) FormattedGoString(depth int) string {
	if this == nil {
		return "nil"
	}
	var s string

	s += getIndentation(depth) + `service ` + this.GetName() + ` {` + "\n"

	// Service Options
	options := this.GetOptions()
	if options != nil {
		s += getFormattedOptionsFromExtensionMap(options.ExtensionMap(), depth, false)
	}

	// Methods	
	if len(this.GetMethod()) > 0 {
		s += "\n"
	}
	for _, method := range this.GetMethod() {
		s += getIndentation(depth+1) + `rpc ` + method.GetName() + `(`
		if len(method.GetInputType()) > 0 {
			s += getLastWordFromPath(method.GetInputType(), ".")
		}
		s += `)`
		if len(method.GetOutputType()) > 0 {
			s += ` returns(` + getLastWordFromPath(method.GetOutputType(), ".") + `)`
		}
		s += " {\n"

		methodOptions := method.GetOptions()
		s += getFormattedOptionsFromExtensionMap(methodOptions.ExtensionMap(), depth+1, false)

		s += getIndentation(depth+1) + "}\n"

	}
	s += getIndentation(depth) + "}"

	return s
}

func getFormattedOptionsFromExtensionMap(extensionMap map[int32]proto.Extension, depth int, fieldOption bool) string {
	var s string
	counter := 0
	if len(extensionMap) > 0 {
		for optInd := range extensionMap {
			// Loop through all imported files
			for _, curFile := range allFiles {
				extensions := curFile.GetExtension()

				// Loop through extensions in the FileDescriptorProto
				for _, ext := range extensions {
					//fmt.Println(ext.GetName())
					if ext.GetNumber() == optInd {
						bytes, _ := proto.GetRawExtension(extensionMap, optInd)
						key, n := proto.DecodeVarint(bytes)

						wt := key & 0x7

						//fmt.Println(fieldDescriptorProtoType_StringValue(ext.GetType()))
						//fmt.Printf("%v\n", wt)
						//fmt.Printf("%v \n", d)

						var val string

						// Enums are special
						if ext.GetType() == FieldDescriptorProto_TYPE_ENUM {
							// Loop through enums to find right one
							for _, myEnum := range curFile.GetEnumType() {
								if myEnum.GetName() == getLastWordFromPath(ext.GetTypeName(), ".") {
									for _, enumVal := range myEnum.GetValue() {
										d, _ := proto.DecodeVarint(bytes[n:])

										if uint64(enumVal.GetNumber()) == d {
											val = enumVal.GetName()
										}
									}
								}
							}

							if !fieldOption {
								s += getIndentation(depth + 1)
								s += `option (`
							} else {
								if counter >= 1 {
									s += ", "
								}
								s += `(`
							}

							if curFile.GetName() != currentFile.GetName() {
								s += curFile.GetPackage() + "."
							}
							s += ext.GetName() + ") = " + val

							if !fieldOption {
								s += ";\n"
							}
							counter += 1

						} else if wt == 2 && ext.GetType() != FieldDescriptorProto_TYPE_STRING { // Messages are special (for method options)
							payload, m := proto.DecodeVarint(bytes[n:])
							n += m
							//fmt.Printf("payload: %v\n", payload)
							for ind := uint64(0); ind < payload-1; ind += 1 {
								firstInt, a := proto.DecodeVarint(bytes[n:])
								n += a
								wiretype := firstInt & 0x7
								var packType FieldDescriptorProto_Type
								switch wiretype {
								case 0:
									packType = FieldDescriptorProto_TYPE_INT32

								case 1:
									packType = FieldDescriptorProto_TYPE_FIXED64

								case 2:
									packType = FieldDescriptorProto_TYPE_STRING

								case 5:
									packType = FieldDescriptorProto_TYPE_FIXED32
								}

								packVal, b := byteToValueString(bytes, n, packType)
								n += b
								//fmt.Printf("%v\n", packVal)

								if ext.GetType() == FieldDescriptorProto_TYPE_MESSAGE {
									tagNum := firstInt >> 3
									//fmt.Printf("tagnum: %v\n", tagNum)

									// Loop through messages to find right one
									myMessage := curFile.GetMessage(getLastWordFromPath(ext.GetTypeName(), "."))
									//fmt.Println(myMessage.GetName())
									for _, field := range myMessage.GetField() {
										if uint64(field.GetNumber()) == tagNum {
											val = `.` + field.GetName() + " = " + packVal
										}
									}
								}
								if !fieldOption {
									s += getIndentation(depth + 1)
									s += `option (`
								} else {
									if counter >= 1 {
										s += ", "
									}
									s += `(`
								}

								if curFile.GetName() != currentFile.GetName() {
									s += curFile.GetPackage() + "."
								}
								s += ext.GetName() + ")" + val
								if !fieldOption {
									s += ";\n"
								}

								counter += 1

							}

						} else {
							val, _ = byteToValueString(bytes, n, ext.GetType())

							if !fieldOption {
								s += getIndentation(depth + 1)
								s += `option (`
							} else {
								if counter >= 1 {
									s += ", "
								}
								s += `(`
							}

							if curFile.GetName() != currentFile.GetName() {
								s += curFile.GetPackage() + "."
							}
							s += ext.GetName() + ") = " + val

							if !fieldOption {
								s += ";\n"
							}
							counter += 1
						}

					}
				}
			}
		}
	}

	return s
}

// Determines depth of indentation
func getIndentation(depth int) string {
	s := ""
	for i := 0; i < depth; i++ {
		s += INDENT
	}
	return s
}

// returns the string representation of a field label
func fieldDescriptorProtoLabel_StringValue(label FieldDescriptorProto_Label) string {
	switch label {
	case FieldDescriptorProto_LABEL_OPTIONAL:
		return "optional"
	case FieldDescriptorProto_LABEL_REQUIRED:
		return "required"
	case FieldDescriptorProto_LABEL_REPEATED:
		return "repeated"
	}

	return "nil"
}

// returns the string representation of a field type
func fieldDescriptorProtoType_StringValue(fieldType FieldDescriptorProto_Type) string {
	switch fieldType {
	case FieldDescriptorProto_TYPE_DOUBLE:
		return "double"
	case FieldDescriptorProto_TYPE_FLOAT:
		return "float"
	case FieldDescriptorProto_TYPE_INT64:
		return "int64"
	case FieldDescriptorProto_TYPE_UINT64:
		return "uint64"
	case FieldDescriptorProto_TYPE_INT32:
		return "int32"
	case FieldDescriptorProto_TYPE_FIXED64:
		return "fixed64"
	case FieldDescriptorProto_TYPE_FIXED32:
		return "fixed32"
	case FieldDescriptorProto_TYPE_BOOL:
		return "bool"
	case FieldDescriptorProto_TYPE_STRING:
		return "string"
	case FieldDescriptorProto_TYPE_GROUP:
		return "group"
	case FieldDescriptorProto_TYPE_MESSAGE:
		return "message"
	case FieldDescriptorProto_TYPE_BYTES:
		return "bytes"
	case FieldDescriptorProto_TYPE_UINT32:
		return "uint32"
	case FieldDescriptorProto_TYPE_ENUM:
		return "enum"
	case FieldDescriptorProto_TYPE_SFIXED32:
		return "sfixed32"
	case FieldDescriptorProto_TYPE_SFIXED64:
		return "sfixed64"
	case FieldDescriptorProto_TYPE_SINT32:
		return "sint32"
	case FieldDescriptorProto_TYPE_SINT64:
		return "sint64"
	}

	return "nil"
}

func byteToValueString(b []byte, lastReadIndex int, t FieldDescriptorProto_Type) (string, int) {
	var val string
	// All the types of options
	switch t {
	case FieldDescriptorProto_TYPE_BOOL:
		d, _ := proto.DecodeVarint(b[lastReadIndex:])
		if int(d) == 1 {
			val = "true"
		} else {
			val = "false"
		}
		lastReadIndex += 1

	case FieldDescriptorProto_TYPE_UINT32, FieldDescriptorProto_TYPE_INT32, FieldDescriptorProto_TYPE_UINT64, FieldDescriptorProto_TYPE_INT64:
		d, _ := proto.DecodeVarint(b[lastReadIndex:])
		lastReadIndex += 1
		val = fmt.Sprintf("%v", d)

	case FieldDescriptorProto_TYPE_FLOAT, FieldDescriptorProto_TYPE_SFIXED32, FieldDescriptorProto_TYPE_FIXED32:
		var f float32
		binary.Read(bytes.NewBuffer(b[lastReadIndex:]), binary.LittleEndian, &f)
		lastReadIndex += 1
		val = fmt.Sprintf("%v", f)

	case FieldDescriptorProto_TYPE_DOUBLE, FieldDescriptorProto_TYPE_SFIXED64, FieldDescriptorProto_TYPE_FIXED64:
		var f float64
		binary.Read(bytes.NewBuffer(b[lastReadIndex:]), binary.LittleEndian, &f)
		lastReadIndex += 1
		val = fmt.Sprintf("%v", f)

	case FieldDescriptorProto_TYPE_STRING:
		_, n := proto.DecodeVarint(b[lastReadIndex:])
		lastReadIndex += n
		val = `"` + string(b[lastReadIndex:]) + `"`

	}

	return val, lastReadIndex
}

func getLastWordFromPath(s string, d string) string {
	return strings.Split(s, d)[len(strings.Split(s, d))-1]
}
