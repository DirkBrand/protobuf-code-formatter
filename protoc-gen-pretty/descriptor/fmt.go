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
	"encoding/binary"
	fmt "fmt"
	regex "regexp"
	sort "sort"
	strings "strings"

	proto "github.com/gogo/protobuf/proto"
)

const (
	INDENT = "  "
)

var importsList []string
var allFiles []*FileDescriptor
var thisFile *FileDescriptor

var commentsMap map[string]*SourceCodeInfo_Location

// Handles the set of Files (but for provided filename only)
func (this *FileDescriptorSet) Fmt(fileToFormat string) string {
	// Loop through all the FileDescriptorProto
	allFiles = make([]*FileDescriptor, len(this.File))
	WrapTypes(this)
	for _, tmpFile := range allFiles {
		if tmpFile.GetName() == fileToFormat {
			thisFile = tmpFile
			s := tmpFile.Fmt(0)
			//fmt.Println(tmpFile.GoString())
			s = strings.Replace(s, "\n\n\n", "\n\n", -1)
			return s
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

			lc := LeadingComments(fmt.Sprintf("%d,%d", importPath, ind), depth)
			if len(lc) > 0 {
				if ind == 0 {
					s = append(s, strings.TrimPrefix(lc, "\n"))
				} else {
					s = append(s, lc)
				}
			}
			s = append(s, `import "`)
			s = append(s, imp)
			s = append(s, `";`)
			s = append(s, "\n")
			s = append(s, TrailingComments(fmt.Sprintf("%d,%d", importPath, ind), depth))
		}

		counter += 1
	}

	// Special options
	options := this.GetOptions()
	optionCount := 0
	var optSlice []string
	if options != nil {
		if (len(this.GetOptions().GetJavaPackage()) > 0 ||
			len(this.GetOptions().GetJavaOuterClassname()) != 0 ||
			this.GetOptions().GetJavaMultipleFiles() ||
			this.GetOptions().GetJavaGenerateEqualsAndHash() ||
			int32(*this.GetOptions().GetOptimizeFor().Enum()) > 1) && counter > 0 {
			s = append(s, "\n")
		}

		// JAVA PACKAGE
		if len(this.GetOptions().GetJavaPackage()) != 0 {
			var singOpt []string
			lc := LeadingComments(fmt.Sprintf("%d,999,%d", optionsPath, optionCount), depth)
			if len(lc) > 0 {
				if optionCount == 0 {
					singOpt = append(singOpt, strings.TrimPrefix(lc, "\n"))
				} else {
					singOpt = append(singOpt, lc)
				}
			}
			singOpt = append(singOpt, "option java_package = ")
			singOpt = append(singOpt, `"`+this.GetOptions().GetJavaPackage()+`"`)
			singOpt = append(singOpt, ";\n")
			singOpt = append(singOpt, TrailingComments(fmt.Sprintf("%d,999,%d", optionsPath, optionCount), depth))
			optionCount += 1

			optSlice = append(optSlice, strings.Join(singOpt, ""))
		}

		// JAVA OUTER CLASSNAME
		if len(this.GetOptions().GetJavaOuterClassname()) != 0 {
			var singOpt []string
			lc := LeadingComments(fmt.Sprintf("%d,999,%d", optionsPath, optionCount), depth)
			if len(lc) > 0 {
				if optionCount == 0 {
					singOpt = append(singOpt, strings.TrimPrefix(lc, "\n"))
				} else {
					singOpt = append(singOpt, lc)
				}
			}
			singOpt = append(singOpt, "option java_outer_classname = ")
			singOpt = append(singOpt, `"`+this.GetOptions().GetJavaOuterClassname()+`"`)
			singOpt = append(singOpt, ";\n")
			singOpt = append(singOpt, TrailingComments(fmt.Sprintf("%d,999,%d", optionsPath, optionCount), depth))
			optionCount += 1
			optSlice = append(optSlice, strings.Join(singOpt, ""))
		}

		// JAVA MULTIPLE FILES
		if this.GetOptions().GetJavaMultipleFiles() {
			var singOpt []string
			lc := LeadingComments(fmt.Sprintf("%d,999,%d", optionsPath, optionCount), depth)
			if len(lc) > 0 {
				if optionCount == 0 {
					singOpt = append(singOpt, strings.TrimPrefix(lc, "\n"))
				} else {
					singOpt = append(singOpt, lc)
				}
			}
			singOpt = append(singOpt, "option java_multiple_files = true")
			singOpt = append(singOpt, ";\n")
			singOpt = append(singOpt, TrailingComments(fmt.Sprintf("%d,999,%d", optionsPath, optionCount), depth))
			optionCount += 1
			optSlice = append(optSlice, strings.Join(singOpt, ""))
		}

		// JAVA GENERATE EQUALS AND HASH
		if this.GetOptions().GetJavaGenerateEqualsAndHash() {
			var singOpt []string
			lc := LeadingComments(fmt.Sprintf("%d,999,%d", optionsPath, optionCount), depth)
			if len(lc) > 0 {
				if optionCount == 0 {
					singOpt = append(singOpt, strings.TrimPrefix(lc, "\n"))
				} else {
					singOpt = append(singOpt, lc)
				}
			}
			singOpt = append(singOpt, "option java_generate_equals_and_hash = true")
			singOpt = append(singOpt, ";\n")
			singOpt = append(singOpt, TrailingComments(fmt.Sprintf("%d,999,%d", optionsPath, optionCount), depth))
			optionCount += 1
			optSlice = append(optSlice, strings.Join(singOpt, ""))
		}

		// GO PACKAGE
		if len(this.GetOptions().GetGoPackage()) > 0 {
			var singOpt []string
			lc := LeadingComments(fmt.Sprintf("%d,999,%d", optionsPath, optionCount), depth)
			if len(lc) > 0 {
				if optionCount == 0 {
					singOpt = append(singOpt, strings.TrimPrefix(lc, "\n"))
				} else {
					singOpt = append(singOpt, lc)
				}
			}
			singOpt = append(singOpt, `option go_package = "`)
			singOpt = append(singOpt, this.GetOptions().GetGoPackage())
			singOpt = append(singOpt, `";`)
			singOpt = append(singOpt, "\n")
			singOpt = append(singOpt, TrailingComments(fmt.Sprintf("%d,999,%d", optionsPath, optionCount), depth))
			optionCount += 1
			optSlice = append(optSlice, strings.Join(singOpt, ""))
		}

		//CC GENERIC SERVICE
		if this.GetOptions().GetCcGenericServices() {
			var singOpt []string
			lc := LeadingComments(fmt.Sprintf("%d,999,%d", optionsPath, optionCount), depth)
			if len(lc) > 0 {
				if optionCount == 0 {
					singOpt = append(singOpt, strings.TrimPrefix(lc, "\n"))
				} else {
					singOpt = append(singOpt, lc)
				}
			}
			singOpt = append(singOpt, "option cc_generic_services = true")
			singOpt = append(singOpt, ";\n")
			singOpt = append(singOpt, TrailingComments(fmt.Sprintf("%d,999,%d", optionsPath, optionCount), depth))
			optionCount += 1
			optSlice = append(optSlice, strings.Join(singOpt, ""))
		}

		//JAVA GENERIC SERVICE
		if this.GetOptions().GetJavaGenericServices() {
			var singOpt []string
			lc := LeadingComments(fmt.Sprintf("%d,999,%d", optionsPath, optionCount), depth)
			if len(lc) > 0 {
				if optionCount == 0 {
					singOpt = append(singOpt, strings.TrimPrefix(lc, "\n"))
				} else {
					singOpt = append(singOpt, lc)
				}
			}
			singOpt = append(singOpt, "option java_generic_services = true")
			singOpt = append(singOpt, ";\n")
			singOpt = append(singOpt, TrailingComments(fmt.Sprintf("%d,999,%d", optionsPath, optionCount), depth))
			optionCount += 1
			optSlice = append(optSlice, strings.Join(singOpt, ""))
		}

		// PY GENERIC SERVICE
		if this.GetOptions().GetPyGenericServices() {
			var singOpt []string
			lc := LeadingComments(fmt.Sprintf("%d,999,%d", optionsPath, optionCount), depth)
			if len(lc) > 0 {
				if optionCount == 0 {
					singOpt = append(singOpt, strings.TrimPrefix(lc, "\n"))
				} else {
					singOpt = append(singOpt, lc)
				}
			}
			singOpt = append(singOpt, "option py_generic_services = true")
			singOpt = append(singOpt, ";\n")
			singOpt = append(singOpt, TrailingComments(fmt.Sprintf("%d,999,%d", optionsPath, optionCount), depth))
			optionCount += 1
			optSlice = append(optSlice, strings.Join(singOpt, ""))
		}

		//OPTIMIZE FOR
		if this.GetOptions().OptimizeFor != nil {
			var singOpt []string
			lc := LeadingComments(fmt.Sprintf("%d,999,%d", optionsPath, optionCount), depth)
			if len(lc) > 0 {
				if optionCount == 0 {
					singOpt = append(singOpt, strings.TrimPrefix(lc, "\n"))
				} else {
					singOpt = append(singOpt, lc)
				}
			}
			singOpt = append(singOpt, "option optimize_for = ")
			if int32(*this.GetOptions().GetOptimizeFor().Enum()) > 1 {
				singOpt = append(singOpt, this.GetOptions().GetOptimizeFor().String())
			} else {
				singOpt = append(singOpt, "SPEED")
			}
			singOpt = append(singOpt, ";\n")
			singOpt = append(singOpt, TrailingComments(fmt.Sprintf("%d,999,%d", optionsPath, optionCount), depth))

			optSlice = append(optSlice, strings.Join(singOpt, ""))
		}
	}
	if len(optSlice) > 0 {
		s = append(s, strings.Join(sortSpecialOptions(optSlice), ""))
	}

	// File Options
	if options != nil && len(options.ExtensionMap()) > 0 {
		s = append(s, "\n")
		theOption := getFormattedOptionsFromExtensionMap(options.ExtensionMap(), -1, false, fmt.Sprintf("%d", optionsPath), optionCount)

		theOption = sortOptions(theOption)

		s = append(s, strings.Join(theOption, ""))

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
			s = append(s, strings.TrimPrefix(LeadingComments(fmt.Sprintf("%d", extendPath), depth), "\n"))
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
	if counter > 0 && len(this.desc) > 0 {
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
		if strings.HasPrefix(i, ".") {
			i = i[1:]
		}
		s = append(s, i)
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

		opts := getFormattedOptionsFromExtensionMap(mesOptions.ExtensionMap(), depth, false, fmt.Sprintf("%d,%d", this.path, messageOptionsPath), 0)
		opts = sortOptions(opts)
		s = append(s, strings.Join(opts, ""))
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
			lc := LeadingComments(fmt.Sprintf("%s,%d,%d", this.path, messageFieldPath, i), depth+1)
			if len(lc) > 0 {
				if i == 0 {
					s = append(s, strings.TrimPrefix(lc, "\n"))
				} else {
					s = append(s, lc)
				}
			}

			s = append(s, field.Fmt(depth+1))
			s = append(s, ";")
			tc := TrailingComments(fmt.Sprintf("%s,%d,%d", this.path, messageFieldPath, i), depth+1)
			if len(tc) > 0 {
				s = append(s, tc)
			} else {
				s = append(s, "\n")
			}
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

	s = append(s, "}\n")

	return strings.Join(s, "")
}

// Handles Fields
func (this *FieldDescriptor) Fmt(depth int) string {
	if this == nil {
		return "nil"
	}
	var s []string

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
		if this.parent != nil {
			// Maybe in other another message
			for _, mes := range this.parent.DescriptorProto.GetNestedType() {
				if b, str := scanNestedMessages(mes, typeName, ""); b {
					typeName = str
					found = true
					break
				}
			}
			// Maybe in other enums
			if !found {
				for _, mes := range this.parent.enum {
					if strcmp(mes.GetName(), typeName) == 0 {
						found = true
						break
					}
				}
			}
			// Maybe same package
			if !found {
				for _, curFile := range allFiles {
					// Same Package
					if strcmp(curFile.GetPackage(), currentFile.GetPackage()) == 0 {
						// Look in messages
						for _, mes := range curFile.GetMessageType() {
							if b, str := scanNestedMessages(mes, typeName, ""); b {
								typeName = str
								found = true
								break
							}
						}
						if !found {
							// Look in enums
							for _, enum := range curFile.GetEnumType() {
								if strcmp(enum.GetName(), typeName) == 0 {
									found = true
									break
								}
							}
						}
					}
				}
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
	if options != nil || len(this.GetDefaultValue()) > 0 {
		s = append(s, ` [`)
	}

	if len(this.GetDefaultValue()) > 0 {
		s = append(s, `default=`)
		s = append(s, this.GetDefaultValue())
		i += 1
	}
	if options != nil {
		if options.GetPacked() || options.GetLazy() || options.GetDeprecated() || len(options.ExtensionMap()) > 0 {

			if len(options.ExtensionMap()) > 0 {
				if i >= 1 {
					s = append(s, ", ")
				} else {
					i += 1
				}
				opts := getFormattedOptionsFromExtensionMap(options.ExtensionMap(), -1, true, "", 0)
				s = append(s, strings.Join(opts, ""))
			}

			if options.GetPacked() {
				if i >= 1 {
					s = append(s, ", ")
				}
				s = append(s, `packed=true`)
				i += 1
			}

			if options.GetLazy() {
				if i >= 1 {
					s = append(s, ", ")
				}
				s = append(s, `lazy=true`)
				i += 1
			}

			if options.GetDeprecated() {
				if i >= 1 {
					s = append(s, ", ")
				}
				s = append(s, `deprecated=true`)
				i += 1
			}
		}
		if i == 0 {
			s = append(s, `deprecated=false`)
		}
	}
	if options != nil || len(this.GetDefaultValue()) > 0 {
		s = append(s, `]`)
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

		opts := getFormattedOptionsFromExtensionMap(options.ExtensionMap(), depth, false, fmt.Sprintf("%d,%d", this.path, enumOptionsPath), 0)
		opts = sortOptions(opts)
		s = append(s, strings.Join(opts, ""))
	}

	// enum fields
	if len(this.GetValue()) > 0 {
		s = append(s, "\n")
	}
	for i, enumValue := range this.GetValue() {
		// Comments of the enum fields
		lc := LeadingComments(fmt.Sprintf("%s,%d,%d", this.path, enumValuePath, i), depth+1)
		if len(lc) > 0 {
			if i == 0 {
				s = append(s, strings.TrimPrefix(lc, "\n"))
			} else {
				s = append(s, lc)
			}
		}

		s = append(s, getIndentation(depth+1))
		s = append(s, enumValue.GetName())

		s = append(s, ` = `)
		s = append(s, fmt.Sprintf("%v", enumValue.GetNumber()))

		// OPTIONS
		valueOptions := enumValue.GetOptions()
		if valueOptions != nil {
			s = append(s, ` [`)
			opts := getFormattedOptionsFromExtensionMap(valueOptions.ExtensionMap(), -1, true, fmt.Sprintf("%d,%d", this.path, enumValueOptionsPath), 0)
			s = append(s, strings.Join(opts, ""))
			s = append(s, `]`)
		}

		s = append(s, ";")
		tc := TrailingComments(fmt.Sprintf("%s,%d,%d", this.path, enumValuePath, i), 0)
		if len(tc) > 0 {
			s = append(s, " "+tc)
		} else {
			s = append(s, "\n")
		}
	}

	s = append(s, getIndentation(depth))
	s = append(s, "};\n")

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
		opts := getFormattedOptionsFromExtensionMap(options.ExtensionMap(), depth, false, fmt.Sprintf("%s,%d", this.path, serviceOptionsPath), 0)
		opts = sortOptions(opts)
		s = append(s, strings.Join(opts, ""))
	}

	// Methods
	if len(this.GetMethod()) > 0 {
		s = append(s, "\n")
	}
	for i, method := range this.GetMethod() {
		lc := LeadingComments(fmt.Sprintf("%s,%d,%d", this.path, methodDescriptorPath, i), depth+1)
		if len(lc) > 0 {
			if i == 0 {
				s = append(s, strings.TrimPrefix(lc, "\n"))
			} else {
				s = append(s, lc)
			}
		}
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
		opts := getFormattedOptionsFromExtensionMap(method.GetOptions().ExtensionMap(), depth+1, false, fmt.Sprintf("%s,%d,%d,%d", this.path, methodDescriptorPath, i, methodOptionsPath), 0)
		s = append(s, strings.Join(opts[1:], ""))

		s = append(s, getIndentation(depth+1))
		s = append(s, "}\n")

	}
	s = append(s, getIndentation(depth))
	s = append(s, "}")

	return strings.Join(s, "")
}

func getFormattedOptionsFromExtensionMap(extensionMap map[int32]proto.Extension, depth int, fieldOption bool, pathIncludingParent string, startIndex int) []string {
	var s []string
	counter := 0
	if len(extensionMap) > 0 {
		commentsIndex := startIndex
		// Sort extension map
		for optInd := range extensionMap {
			// Loop through all imported files
			for _, curFile := range allFiles {
				extensions := curFile.GetExtension()

				// Loop through extensions in the FileDescriptorProto
				for ext_i, ext := range extensions {
					if ext.GetNumber() == optInd {
						bytes, _ := proto.GetRawExtension(extensionMap, optInd)
						key, n := proto.DecodeVarint(bytes)
						headerlength := n

						wt := key & 0x7

						var val string

						var singleOption []string

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

							lc := LeadingComments(fmt.Sprintf("%s,%d,%d", pathIncludingParent, 999, commentsIndex), depth+1)
							if len(lc) > 0 {
								if ext_i == 0 {
									singleOption = append(singleOption, strings.TrimPrefix(lc, "\n"))
								} else {
									singleOption = append(singleOption, lc)
								}
							}
							if !fieldOption {
								singleOption = append(singleOption, getIndentation(depth+1))
								singleOption = append(singleOption, `option (`)
							} else {
								if counter >= 1 {
									singleOption = append(singleOption, ", ")
								}
								singleOption = append(singleOption, `(`)
							}

							if curFile.GetName() != currentFile.GetName() && len(curFile.GetPackage()) > 0 {

								singleOption = append(singleOption, curFile.GetPackage())
								singleOption = append(singleOption, ".")
							}
							singleOption = append(singleOption, ext.GetName())
							singleOption = append(singleOption, ")=")
							singleOption = append(singleOption, val)

							if !fieldOption {
								singleOption = append(singleOption, ";\n")
							}
							comm := TrailingComments(fmt.Sprintf("%s,%d,%d", pathIncludingParent, 999, commentsIndex), depth+1)
							if len(comm) > 0 {
								singleOption = append(singleOption, comm)
								if counter < len(extensionMap)-1 {
									singleOption = append(singleOption, "\n")
								}
							}
							s = append(s, strings.Join(singleOption, ""))

							commentsIndex += 1
							counter += 1

						} else if wt == 2 && ext.GetType() != FieldDescriptorProto_TYPE_STRING { // Messages are special (for method options)

							for n < len(bytes) {
								// Grab the payload
								_, m := proto.DecodeVarint(bytes[n:])
								n += m
								// Grab tag/wiretype
								//fmt.Printf("%v, %v\n", n, len(bytes))
								firstInt, a := proto.DecodeVarint(bytes[n:])
								//fmt.Printf("firstint: %v\n", a)
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
								n = b
								//fmt.Printf("%v\n", packType)
								//fmt.Printf("%v\n", packVal)
								n += headerlength

								if ext.GetType() == FieldDescriptorProto_TYPE_MESSAGE {
									tagNum := firstInt >> 3
									//fmt.Printf("tagnum: %v\n", tagNum)

									// Loop through messages to find right one
									myMessage := curFile.GetMessage(getLastWordFromPath(ext.GetTypeName(), "."))
									//fmt.Println(myMessage.GetName())
									for _, field := range myMessage.GetField() {
										if uint64(field.GetNumber()) == tagNum {
											val = `.` + field.GetName() + " = " + packVal
											break
										}
									}
								}

								singleOption = append(singleOption, LeadingComments(fmt.Sprintf("%s,%d,%d", pathIncludingParent, 999, commentsIndex), depth+1))

								if !fieldOption {
									singleOption = append(singleOption, getIndentation(depth+1))
									singleOption = append(singleOption, `option (`)
								} else {
									if counter >= 1 {
										singleOption = append(singleOption, ", ")
									}
									singleOption = append(singleOption, `(`)
								}

								if curFile.GetName() != currentFile.GetName() && len(curFile.GetPackage()) > 0 {
									singleOption = append(singleOption, curFile.GetPackage())
									singleOption = append(singleOption, ".")
								}
								singleOption = append(singleOption, ext.GetName())
								singleOption = append(singleOption, ")")
								singleOption = append(singleOption, val)
								if !fieldOption {
									singleOption = append(singleOption, ";\n")
								}
								comm := TrailingComments(fmt.Sprintf("%s,%d,%d", pathIncludingParent, 999, commentsIndex), depth+1)
								if len(comm) > 0 {
									singleOption = append(singleOption, comm)
									if counter < len(extensionMap) {
										singleOption = append(singleOption, "\n")
									}
								}
								commentsIndex += 1
								counter += 1

								s = append(s, strings.Join(singleOption, ""))

							}

						} else {
							val, b := byteToValueString(bytes, n, ext.GetType())
							n = b

							singleOption = append(singleOption, LeadingComments(fmt.Sprintf("%s,%d,%d", pathIncludingParent, 999, commentsIndex), depth+1))

							if !fieldOption {
								singleOption = append(singleOption, getIndentation(depth+1))
								singleOption = append(singleOption, `option (`)
							} else {
								if counter >= 1 {
									singleOption = append(singleOption, ", ")
								}
								singleOption = append(singleOption, `(`)
							}

							if curFile.GetName() != currentFile.GetName() && len(curFile.GetPackage()) > 0 {
								singleOption = append(singleOption, curFile.GetPackage())
								singleOption = append(singleOption, ".")
							}
							singleOption = append(singleOption, ext.GetName())
							singleOption = append(singleOption, ")=")
							singleOption = append(singleOption, val)

							if !fieldOption {
								singleOption = append(singleOption, ";\n")
							}
							comm := TrailingComments(fmt.Sprintf("%s,%d,%d", pathIncludingParent, 999, commentsIndex), depth+1)
							if len(comm) > 0 {
								singleOption = append(singleOption, comm)
								if counter < len(extensionMap) {
									singleOption = append(singleOption, "\n")
								}
							}
							commentsIndex += 1
							counter += 1

							s = append(s, strings.Join(singleOption, ""))
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
		d, m := proto.DecodeVarint(b[lastReadIndex:])
		lastReadIndex += m
		//fmt.Printf("m: %v\n", m)
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
		l, n := proto.DecodeVarint(b[lastReadIndex:])
		lastReadIndex += n
		val = `"` + string(b[lastReadIndex:lastReadIndex+int(l)]) + `"`
		lastReadIndex += int(l)

	}

	return val, lastReadIndex
}

func getLastWordFromPath(s string, d string) string {
	return strings.Split(s, d)[len(strings.Split(s, d))-1]
}

func scanNestedMessages(parent *DescriptorProto, typename string, parentStr string) (bool, string) {
	found := false
	val := ""
	if parent.GetName() == typename {
		var str string
		if len(parentStr) == 0 {
			str = typename
		} else {
			str = parentStr + "." + typename
		}
		return true, str
	} else {
		for _, mes := range parent.GetNestedType() {
			var str string
			if len(parentStr) == 0 {
				str = parent.GetName()
			} else {
				str = parentStr + "." + parent.GetName()
			}
			b, val := scanNestedMessages(mes, typename, str)

			if b {
				return b, val
			}
		}
	}

	return found, val
}

func strcmp(a, b string) int {
	var min = len(b)
	if len(a) < len(b) {
		min = len(a)
	}
	var diff int
	for i := 0; i < min && diff == 0; i++ {
		diff = int(a[i]) - int(b[i])
	}
	if diff == 0 {
		diff = len(a) - len(b)
	}
	return diff
}

func sortSpecialOptions(opts []string) []string {
	var vals []string

	var s map[string]string
	s = make(map[string]string)
	r1, _ := regex.Compile(`option `)
	r2, _ := regex.Compile(`([\s]*=[\s]*)`)
	for _, opt := range opts {
		b1 := r1.FindStringIndex(opt)
		b2 := r2.FindStringIndex(opt)
		s[opt[b1[1]:b2[0]]] = opt
	}

	var keys []string
	for k := range s {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vals = append(vals, s[k])
	}

	return vals
}

func sortOptions(opts []string) []string {
	var vals []string

	var s map[string]string
	s = make(map[string]string)
	r1, _ := regex.Compile(`option [(]`)
	r2, _ := regex.Compile(`[)]([\s]*=[\s]*)`)
	for _, opt := range opts {
		b1 := r1.FindStringIndex(opt)
		b2 := r2.FindStringIndex(opt)
		s[opt[b1[1]:b2[0]]] = opt
	}

	var keys []string
	for k := range s {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vals = append(vals, s[k])
	}

	return vals
}
