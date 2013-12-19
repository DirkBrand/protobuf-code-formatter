/*

Copyright (c) 2013, Dirk Brand
All rights reserved.

Redistribution and use in source and binary forms, with or without modification, are permitted
provided that the following conditions are met:

 * Redistributions of source code must retain the above copyright notice, this list of
   conditions and the following disclaimer.
 * Redistributions in binary form must reproduce the above copyright notice, this list of
   conditions and the following disclaimer in the documentation and/or other materials provided
   with the distribution.

THIS SOFTWARE IS PROVIDED BY THE AUTHOR AND CONTRIBUTORS ``AS IS'' AND ANY EXPRESS OR IMPLIED
WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND
FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE AUTHOR OR CONTRIBUTORS
BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA,
OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT
OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

*/

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

var importsList []string
var allFiles []*FileDescriptor

var commentsMap map[string]*SourceCodeInfo_Location

// Handles the set of Files (but for provided filename only)
func (this *FileDescriptorSet) Fmt(fileToFormat string) string {
	// Loop through all the FileDescriptorProto
	allFiles = make([]*FileDescriptor, len(this.File))
	WrapTypes(this)
	for _, tmpFile := range allFiles {
		if tmpFile.GetName() == fileToFormat {
			return tmpFile.Fmt(0)
		}
	}
	return ""
}

// Handles entire file
func (this *FileDescriptor) Fmt(depth int) string {
	if this == nil {
		return "nil"
	}
	currentFile = *this

	var s []string

	counter := 0

	// SourceCodeInfo
	// Generates a map of comments to paths to use in the construction of the program
	//source := this.GetSourceCodeInfo()
	//fmt.Println(len(source.GetLocation()))
	//commentsMap = source.ExtractComments()

	// the package
	if len(this.GetPackage()) > 0 {
		s = append(s, LeadingComments(fmt.Sprintf("%d", packagePath), depth))
		s = append(s, `package `)
		s = append(s, this.GetPackage())
		s = append(s, ";\n")
		s = append(s, TrailingComments(fmt.Sprintf("%d", packagePath), depth))

		counter += 1
	}

	// For each import
	if len(this.GetDependency()) > 0 && counter > 0 {
		s = append(s, "\n")
	}
	if len(this.GetDependency()) > 0 {
		importsList = make([]string, len(this.GetDependency()))
		i := 0

		sort.Strings(this.GetDependency())
		for ind, imp := range this.GetDependency() {
			importsList[i] = strings.Split(imp, "/")[len(strings.Split(imp, "/"))-1]
			i += 1

			s = append(s, LeadingComments(fmt.Sprintf("%d,%d", importPath, ind), depth))
			s = append(s, `import "`)
			s = append(s, imp)
			s = append(s, `";`)
			s = append(s, "\n")
			s = append(s, TrailingComments(fmt.Sprintf("%d,%d", importPath, ind), depth))
		}

		counter += 1
	}

	// For each extend
	extendGroups := make(map[string]string)
	for i, ext := range this.ext {
		extendGroups[ext.GetExtendee()] = extendGroups[ext.GetExtendee()] + LeadingComments(fmt.Sprintf("%d,%d", extendPath, i), depth+1) + ext.Fmt(depth+1) + ";\n" + TrailingComments(fmt.Sprintf("%d,%d", extendPath, i), depth+1)

	}
	if len(extendGroups) > 0 && counter > 0 {
		s = append(s, "\n")
	}
	ind := 0
	for i := range extendGroups {
		group := extendGroups[i]
		if ind == 0 {
			s = append(s, LeadingComments(fmt.Sprintf("%d", extendPath), depth))
		} else {
			s = append(s, "\n")
			s = append(s, LeadingComments(fmt.Sprintf("%d,%d", extendPath, ind*1000), depth))
		}
		s = append(s, getIndentation(depth))
		s = append(s, `extend `)
		s = append(s, strings.Replace(i, ".", "", 1))
		s = append(s, " {\n")
		s = append(s, group)
		s = append(s, getIndentation(depth))
		s = append(s, "}\n")

		if ind == 0 {
			s = append(s, TrailingComments(fmt.Sprintf("%d", extendPath), depth))
		} else {
			s = append(s, TrailingComments(fmt.Sprintf("%d,%d", extendPath, ind*1000), depth))
		}

		ind += 1
		counter += 1
	}

	// File Options
	options := this.GetOptions()
	if options != nil && len(options.ExtensionMap()) > 0 {
		s = append(s, "\n")
		theOption := getFormattedOptionsFromExtensionMap(options.ExtensionMap(), -1, false, fmt.Sprintf("%d", optionsPath))
		s = append(s, theOption)

		counter += 1
	}

	// Enums
	if len(this.enum) > 0 && counter > 0 {
		s = append(s, "\n")
	}
	for _, enum := range this.enum {
		s = append(s, enum.Fmt(depth))
		s = append(s, "\n")

		counter += 1
	}

	// Messages
	if len(this.desc) > 0 && counter > 0 {
		s = append(s, "\n")
	}
	for _, message := range this.desc {
		if message.parent == nil {
			s = append(s, message.Fmt(depth, false, nil))
			s = append(s, "\n")

			counter += 1
		}
	}

	// Services
	if len(this.serv) > 0 && counter > 0 {
		s = append(s, "\n")
	}
	for _, service := range this.serv {
		s = append(s, service.Fmt(depth))
		s = append(s, "\n")

		counter += 1
	}

	return strings.Join(s, "")
}

// Handles Messages
func (this *Descriptor) Fmt(depth int, isGroup bool, groupField *FieldDescriptorProto) string {
	if this == nil {
		return "nil"
	}
	var s []string
	contentCount := 0

	// For groups logic
	nestedMessages := this.nested

	// Message Header
	s = append(s, LeadingComments(this.path, depth))

	if isGroup {
		s = append(s, getIndentation(depth))
		s = append(s, fieldDescriptorProtoLabel_StringValue(*groupField.Label))
		s = append(s, ` group `)
		s = append(s, this.GetName())
		s = append(s, " = ")
		s = append(s, fmt.Sprintf("%v", *groupField.Number))
		s = append(s, ` {`)
	} else {
		s = append(s, getIndentation(depth))
		s = append(s, `message `)
		s = append(s, this.GetName())
		s = append(s, ` {`)
	}
	tc := TrailingComments(this.path, depth+1)
	if len(tc) > 0 {
		s = append(s, "\n")
		s = append(s, tc)
	}

	// Extension Range
	if len(this.GetExtensionRange()) > 0 {
		s = append(s, "\n")
		contentCount += 1
	}
	for extensionIndex, ext := range this.GetExtensionRange() {
		if extensionIndex == 0 {
			s = append(s, LeadingComments(fmt.Sprintf("%s,%d", this.path, messageExtensionRangePath), depth+1))
		} else {
			s = append(s, "\n")
			s = append(s, LeadingComments(fmt.Sprintf("%s,%d,%d", this.path, messageExtensionRangePath, extensionIndex*1000), depth+1))
		}
		s = append(s, ext.Fmt(depth+1))

		if extensionIndex == 0 {
			s = append(s, TrailingComments(fmt.Sprintf("%s,%d", this.path, messageExtensionRangePath), depth+1))
		} else {
			s = append(s, "\n")
			s = append(s, TrailingComments(fmt.Sprintf("%s,%d,%d", this.path, messageExtensionRangePath, extensionIndex*1000), depth+1))
		}
	}

	// For each extend
	if len(this.ext) > 0 {
		s = append(s, "\n")
		contentCount += 1
	}
	extendGroups := make(map[string]string)
	for index, ext := range this.ext {
		if depth == 0 {
			extendGroups[ext.GetExtendee()] = extendGroups[ext.GetExtendee()] + LeadingComments(fmt.Sprintf("%s,%d,%d", this.path, messageExtensionPath, index), depth+2) + getIndentation(depth+1) + ext.Fmt(depth+1) + ";\n" + TrailingComments(fmt.Sprintf("%s,%d,%d", this.path, messageExtensionPath, index), depth+2)
		} else {
			extendGroups[ext.GetExtendee()] = extendGroups[ext.GetExtendee()] + LeadingComments(fmt.Sprintf("%s,%d,%d", this.path, messageExtensionPath, index), depth+2) + ext.Fmt(depth+1) + ";\n" + TrailingComments(fmt.Sprintf("%s,%d,%d", this.path, messageExtensionPath, index), depth+2)
		}

	}
	index := 0
	for i := range extendGroups {
		group := extendGroups[i]
		if index == 0 {
			s = append(s, LeadingComments(fmt.Sprintf("%s,%d", this.path, messageExtensionPath), depth+1))
		} else {
			s = append(s, "\n")
			s = append(s, LeadingComments(fmt.Sprintf("%s,%d,%d", this.path, messageExtensionPath, index*1000), depth+1))
		}
		s = append(s, getIndentation(depth+1))
		s = append(s, `extend `)
		s = append(s, getLastWordFromPath(i, "."))
		s = append(s, " {\n")
		if index == 0 {
			tc := TrailingComments(fmt.Sprintf("%s,%d", this.path, messageExtensionPath), depth+1)
			if len(tc) > 0 {
				s = append(s, getIndentation(depth+1))
				s = append(s, tc)
				s = append(s, "\n")
			}
		} else {
			tc := TrailingComments(fmt.Sprintf("%s,%d,%d", this.path, messageExtensionPath, index*1000), depth+1)
			if len(tc) > 0 {
				s = append(s, getIndentation(depth+1))
				s = append(s, tc)
				s = append(s, "\n")
			}
		}

		s = append(s, group)
		s = append(s, getIndentation(depth+1))
		s = append(s, "}\n")

		index += 1
	}

	// Options
	mesOptions := this.GetOptions()
	if mesOptions != nil && len(mesOptions.ExtensionMap()) > 0 {
		s = append(s, "\n")
		contentCount += 1

		s = append(s, getFormattedOptionsFromExtensionMap(mesOptions.ExtensionMap(), depth, false, fmt.Sprintf("%d,%d", this.path, messageOptionsPath)))
	}

	// Fields
	if len(this.field) > 0 {
		s = append(s, "\n")
		contentCount += 1
	}
	for i, field := range this.field {

		if field.GetType() == FieldDescriptorProto_TYPE_GROUP {
			for i := 0; i < len(nestedMessages); i += 1 {
				nestedMes := nestedMessages[i]
				// Found group
				if strings.ToLower(nestedMes.GetName()) == field.GetName() {
					s = append(s, "\n")
					tempStr := nestedMes.Fmt(depth+1, true, field.FieldDescriptorProto)
					s = append(s, tempStr)
					nestedMessages = append(nestedMessages[:i], nestedMessages[i+1:]...)
				}
			}
		} else {
			s = append(s, LeadingComments(fmt.Sprintf("%s,%d,%d", this.path, messageFieldPath, i), depth+1))
			s = append(s, field.Fmt(depth+1))
			s = append(s, ";\n")
			s = append(s, TrailingComments(fmt.Sprintf("%s,%d,%d", this.path, messageFieldPath, i), depth+1))
		}
	}

	// Enums
	if len(this.GetEnumType()) > 0 {
		s = append(s, "\n")
		contentCount += 1
	}
	for _, enum := range this.enum {
		s = append(s, enum.Fmt(depth+1))
	}

	// Nested Messages
	if len(nestedMessages) > 0 {
		s = append(s, "\n")
		contentCount += 1
	}
	for _, nestedMessage := range nestedMessages {
		s = append(s, nestedMessage.Fmt(depth+1, false, nil))
	}

	if contentCount > 0 {
		s = append(s, getIndentation(depth))
	}
	s = append(s, "}\n")

	return strings.Join(s, "")
}

// Handles Fields
func (this *FieldDescriptor) Fmt(depth int) string {
	if this == nil {
		return "nil"
	}
	var s []string

	//s = append(s, LeadingComments(this.path))
	s = append(s, getIndentation(depth))
	s = append(s, fieldDescriptorProtoLabel_StringValue(*this.Label))
	s = append(s, ` `)
	// If referencing a message
	if *this.Type == FieldDescriptorProto_TYPE_MESSAGE || *this.Type == FieldDescriptorProto_TYPE_ENUM {
		var found bool
		typeName := getLastWordFromPath(this.GetTypeName(), ".")
		for _, mes := range currentFile.GetMessageType() {
			if mes.GetName() == typeName {
				found = true
			}
		}
		if found {
			s = append(s, typeName)
		} else {
			typeName = this.GetTypeName()
			if strings.HasPrefix(typeName, ".") {
				typeName = typeName[1:]
			}
			s = append(s, typeName)
		}

	} else {
		s = append(s, fieldDescriptorProtoType_StringValue(*this.Type))
	}
	s = append(s, ` `)
	s = append(s, this.GetName())
	s = append(s, ` = `)
	s = append(s, fmt.Sprintf("%v", this.GetNumber()))

	// OPTIONS
	options := this.GetOptions()
	i := 0
	if options != nil {
		if options.GetPacked() || options.GetLazy() || options.GetDeprecated() || len(this.GetDefaultValue()) > 0 || len(options.ExtensionMap()) > 0 {
			s = append(s, ` [`)

			if len(options.ExtensionMap()) > 0 {
				s = append(s, getFormattedOptionsFromExtensionMap(options.ExtensionMap(), -1, true, ""))
				i += 1
			}

			if len(this.GetDefaultValue()) > 0 {
				if i >= 1 {
					s = append(s, ", ")
				}
				s = append(s, `default = `)
				s = append(s, this.GetDefaultValue())
			}

			if options.GetPacked() {
				if i >= 1 {
					s = append(s, ", ")
				}
				s = append(s, `packed = true`)
			}

			if options.GetLazy() {
				if i >= 1 {
					s = append(s, ", ")
				}
				s = append(s, `lazy = true`)
			}

			if options.GetDeprecated() {
				if i >= 1 {
					s = append(s, ", ")
				}
				s = append(s, `deprecated = true`)
			}

			s = append(s, `]`)
		}
	}
	return strings.Join(s, "")
}

// Handles Enums
func (this *EnumDescriptor) Fmt(depth int) string {
	if this == nil {
		return "nil"
	}
	var s []string

	// Comments of the enum
	s = append(s, LeadingComments(this.path, depth))

	s = append(s, getIndentation(depth))
	s = append(s, `enum `)
	s = append(s, this.GetName())
	s = append(s, ` {`)

	tc := TrailingComments(this.path, depth+1)
	if len(tc) > 0 {
		s = append(s, "\n")
		s = append(s, tc)
	}

	// Options
	options := this.GetOptions()
	if options != nil && len(options.ExtensionMap()) > 0 {
		s = append(s, "\n")

		s = append(s, getFormattedOptionsFromExtensionMap(options.ExtensionMap(), depth, false, fmt.Sprintf("%d,%d", this.path, enumOptionsPath)))
	}

	if len(this.GetValue()) > 0 {
		s = append(s, "\n")
	}
	for i, enumValue := range this.GetValue() {

		// Comments of the enum fields
		s = append(s, LeadingComments(fmt.Sprintf("%s,%d,%d", this.path, enumValuePath, i), depth+1))

		s = append(s, getIndentation(depth+1))
		s = append(s, enumValue.GetName())
		s = append(s, ` = `)
		s = append(s, fmt.Sprintf("%v", enumValue.GetNumber()))

		// OPTIONS
		valueOptions := enumValue.GetOptions()
		if valueOptions != nil {
			s = append(s, ` [`)
			s = append(s, getFormattedOptionsFromExtensionMap(valueOptions.ExtensionMap(), -1, true, fmt.Sprintf("%d,%d", this.path, enumValueOptionsPath)))
			s = append(s, `]`)
		}

		s = append(s, ";\n")

		s = append(s, TrailingComments(fmt.Sprintf("%s,%d,%d", this.path, enumValuePath, i), depth+1))
	}

	s = append(s, getIndentation(depth))
	s = append(s, "}\n")

	return strings.Join(s, "")
}

// Handles Extension Ranges
func (this *DescriptorProto_ExtensionRange) Fmt(depth int) string {
	if this == nil {
		return "nil"
	}
	var s []string
	s = append(s, getIndentation(depth))
	s = append(s, `extensions `)
	s = append(s, fmt.Sprintf("%v", this.GetStart()))
	s = append(s, ` to `)
	if this.GetEnd() >= 1<<29-1 {
		s = append(s, "max;\n")
	} else {
		s = append(s, fmt.Sprintf("%v", this.GetEnd()-1))
		s = append(s, ";\n")
	}

	return strings.Join(s, "")
}

// Handles Services
func (this *ServiceDescriptor) Fmt(depth int) string {
	if this == nil {
		return "nil"
	}
	var s []string

	s = append(s, LeadingComments(this.path, depth))
	s = append(s, getIndentation(depth))
	s = append(s, `service `)
	s = append(s, this.GetName())
	s = append(s, ` {`)
	s = append(s, "\n")

	tc := TrailingComments(this.path, depth+1)
	if len(tc) > 0 {
		s = append(s, tc)
		s = append(s, "\n")
	}

	// Service Options
	options := this.GetOptions()
	if options != nil {
		s = append(s, getFormattedOptionsFromExtensionMap(options.ExtensionMap(), depth, false, fmt.Sprintf("%s,%d", this.path, serviceOptionsPath)))
	}

	// Methods
	if len(this.GetMethod()) > 0 {
		s = append(s, "\n")
	}
	for i, method := range this.GetMethod() {
		s = append(s, LeadingComments(fmt.Sprintf("%s,%d,%d", this.path, methodDescriptorPath, i), depth+1))
		s = append(s, getIndentation(depth+1))
		s = append(s, `rpc `)
		s = append(s, method.GetName())
		s = append(s, `(`)
		if len(method.GetInputType()) > 0 {
			s = append(s, getLastWordFromPath(method.GetInputType(), "."))
		}
		s = append(s, `)`)
		if len(method.GetOutputType()) > 0 {
			s = append(s, ` returns(`)
			s = append(s, getLastWordFromPath(method.GetOutputType(), "."))
			s = append(s, `)`)
		}
		s = append(s, " {\n")
		tc := TrailingComments(fmt.Sprintf("%s,%d,%d", this.path, methodDescriptorPath, i), depth+2)
		if len(tc) > 0 {
			s = append(s, tc)
			s = append(s, "\n")
		}

		methodOptions := method.GetOptions()
		s = append(s, getFormattedOptionsFromExtensionMap(methodOptions.ExtensionMap(), depth+1, false, fmt.Sprintf("%s,%d,%d,%d", this.path, methodDescriptorPath, i, methodOptionsPath)))

		s = append(s, getIndentation(depth+1))
		s = append(s, "}\n")

	}
	s = append(s, getIndentation(depth))
	s = append(s, "}")

	return strings.Join(s, "")
}

func getFormattedOptionsFromExtensionMap(extensionMap map[int32]proto.Extension, depth int, fieldOption bool, pathIncludingParent string) string {
	var s []string
	counter := 0
	if len(extensionMap) > 0 {
		commentsIndex := 0
		for optInd := range extensionMap {
			// Loop through all imported files
			for _, curFile := range allFiles {
				extensions := curFile.GetExtension()

				// Loop through extensions in the FileDescriptorProto
				for _, ext := range extensions {
					if ext.GetNumber() == optInd {
						bytes, _ := proto.GetRawExtension(extensionMap, optInd)
						key, n := proto.DecodeVarint(bytes)

						wt := key & 0x7

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

							s = append(s, LeadingComments(fmt.Sprintf("%s,%d,%d", pathIncludingParent, 999, commentsIndex), depth+1))

							if !fieldOption {
								s = append(s, getIndentation(depth+1))
								s = append(s, `option (`)
							} else {
								if counter >= 1 {
									s = append(s, ", ")
								}
								s = append(s, `(`)
							}

							if curFile.GetName() != currentFile.GetName() && len(curFile.GetPackage()) > 0 {
								s = append(s, curFile.GetPackage())
								s = append(s, ".")
							}
							s = append(s, ext.GetName())
							s = append(s, ") = ")
							s = append(s, val)

							if !fieldOption {
								s = append(s, ";\n")
							}
							comm := TrailingComments(fmt.Sprintf("%s,%d,%d", pathIncludingParent, 999, commentsIndex), depth+1)
							if len(comm) > 0 {
								s = append(s, comm)
								if counter < len(extensionMap)-1 {
									s = append(s, "\n")
								}
							}

							commentsIndex += 1
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

								s = append(s, LeadingComments(fmt.Sprintf("%s,%d,%d", pathIncludingParent, 999, commentsIndex), depth+1))

								if !fieldOption {
									s = append(s, getIndentation(depth+1))
									s = append(s, `option (`)
								} else {
									if counter >= 1 {
										s = append(s, ", ")
									}
									s = append(s, `(`)
								}

								if curFile.GetName() != currentFile.GetName() && len(curFile.GetPackage()) > 0 {
									s = append(s, curFile.GetPackage())
									s = append(s, ".")
								}
								s = append(s, ext.GetName())
								s = append(s, ")")
								s = append(s, val)
								if !fieldOption {
									s = append(s, ";\n")
								}
								comm := TrailingComments(fmt.Sprintf("%s,%d,%d", pathIncludingParent, 999, commentsIndex), depth+1)
								if len(comm) > 0 {
									s = append(s, comm)
									if counter < len(extensionMap) {
										s = append(s, "\n")
									}
								}
								commentsIndex += 1
								counter += 1

							}

						} else {
							val, _ = byteToValueString(bytes, n, ext.GetType())

							s = append(s, LeadingComments(fmt.Sprintf("%s,%d,%d", pathIncludingParent, 999, commentsIndex), depth+1))

							if !fieldOption {
								s = append(s, getIndentation(depth+1))
								s = append(s, `option (`)
							} else {
								if counter >= 1 {
									s = append(s, ", ")
								}
								s = append(s, `(`)
							}

							if curFile.GetName() != currentFile.GetName() && len(curFile.GetPackage()) > 0 {
								s = append(s, curFile.GetPackage())
								s = append(s, ".")
							}
							s = append(s, ext.GetName())
							s = append(s, ") = ")
							s = append(s, val)

							if !fieldOption {
								s = append(s, ";\n")
							}
							comm := TrailingComments(fmt.Sprintf("%s,%d,%d", pathIncludingParent, 999, commentsIndex), depth+1)
							if len(comm) > 0 {
								s = append(s, comm)
								if counter < len(extensionMap) {
									s = append(s, "\n")
								}
							}
							commentsIndex += 1
							counter += 1
						}

					}
				}
			}
		}
	}

	return strings.Join(s, "")
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
