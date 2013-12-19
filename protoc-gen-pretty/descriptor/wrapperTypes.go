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
	"fmt"
	"strconv"
	"strings"
)

const (
	// tag numbers in FileDescriptorProto
	packagePath = 2 // package decleration
	importPath  = 3 // imports
	messagePath = 4 // message_type
	enumPath    = 5 // enum_type
	servicePath = 6 // services
	extendPath  = 7 // extensions
	optionsPath = 8 // options

	// tag numbers in DescriptorProto
	messageFieldPath          = 2 // field
	messageMessagePath        = 3 // nested_type
	messageEnumPath           = 4 // enum_type
	messageExtensionRangePath = 5
	messageExtensionPath      = 6
	messageOptionsPath        = 7

	// tag numbers in EnumDescriptorProto
	enumValuePath        = 2 // value
	enumOptionsPath      = 3
	enumValueOptionsPath = 3

	// tag numbers in ServiceDescriptorProto
	methodDescriptorPath = 2
	methodOptionsPath    = 4
	serviceOptionsPath   = 3
)

var currentFile FileDescriptor

type common struct {
	file *FileDescriptorProto // File this object comes from.
}

type Descriptor struct {
	common
	*DescriptorProto
	parent *Descriptor        // The containing message, if any.
	nested []*Descriptor      // Inner messages, if any.
	ext    []*FieldDescriptor // Extensions, if any.
	field  []*FieldDescriptor // Fields, if any.
	enum   []*EnumDescriptor  // Enums, if any.
	index  int                // The index into the container, whether the file or another message.
	path   string             // The SourceCodeInfo path as comma-separated integers.
	group  bool
}

type EnumDescriptor struct {
	common
	*EnumDescriptorProto
	parent *Descriptor // The containing message, if any.
	path   string      // The SourceCodeInfo path as comma-separated integers.
}

type FieldDescriptor struct {
	common
	*FieldDescriptorProto
	parent *Descriptor // The containing message, if any.
}

type ImportedDescriptor struct {
	common
}

type ServiceDescriptor struct {
	common
	*ServiceDescriptorProto
	path string
}

type FileDescriptor struct {
	*FileDescriptorProto
	desc []*Descriptor         // All the messages defined in this file.
	enum []*EnumDescriptor     // All the enums defined in this file.
	ext  []*FieldDescriptor    // All the top-level extensions defined in this file.
	serv []*ServiceDescriptor  // All the top-level services defined in this file.
	imp  []*ImportedDescriptor // All types defined in files publicly imported by this file.

	// Comments, stored as a map of path (comma-separated integers) to the comment.
	comments map[string]*SourceCodeInfo_Location
}

func WrapTypes(set *FileDescriptorSet) {
	for i, f := range set.File {
		// We must wrap the descriptors before we wrap the enums
		descs := wrapDescriptors(f)
		buildNestedDescriptors(descs)
		enums := wrapEnumDescriptors(f, descs)
		exts := wrapExtensions(f)
		serves := wrapServiceDescriptors(f)
		fd := &FileDescriptor{
			FileDescriptorProto: f,
			desc:                descs,
			enum:                enums,
			ext:                 exts,
			serv:                serves,
		}
		extractComments(fd)
		allFiles[i] = fd
	}
}

// Scan the descriptors in this file.  For each one, build the slice of nested descriptors
func buildNestedDescriptors(descs []*Descriptor) {
	for _, desc := range descs {
		if len(desc.NestedType) != 0 {
			desc.nested = make([]*Descriptor, len(desc.NestedType))
			n := 0
			for _, nest := range descs {
				if nest.parent == desc {
					desc.nested[n] = nest
					n++
				}
			}
		}
	}
}

func newDescriptor(desc *DescriptorProto, parent *Descriptor, file *FileDescriptorProto, index int) *Descriptor {
	d := &Descriptor{
		common:          common{file},
		DescriptorProto: desc,
		parent:          parent,
		index:           index,
	}
	if parent == nil {
		d.path = fmt.Sprintf("%d,%d", messagePath, index)
	} else {
		d.path = fmt.Sprintf("%s,%d,%d", parent.path, messageMessagePath, index)
	}

	d.ext = make([]*FieldDescriptor, len(desc.Extension))
	for i, field := range desc.Extension {
		d.ext[i] = &FieldDescriptor{common{file}, field, d}
	}

	d.field = make([]*FieldDescriptor, len(desc.Field))
	for i, field := range desc.Field {
		d.field[i] = &FieldDescriptor{common{file}, field, d}
	}

	// Enums within messages. Enums within embedded messages appear in the outer-most message.
	d.enum = make([]*EnumDescriptor, len(desc.EnumType))
	for i, enums := range desc.GetEnumType() {
		d.enum[i] = newEnumDescriptor(enums, d, d.common.file, i)
	}

	return d
}

// Return a slice of all the Descriptors defined within this file
func wrapDescriptors(file *FileDescriptorProto) []*Descriptor {
	sl := make([]*Descriptor, 0, len(file.MessageType)+10)
	for i, desc := range file.MessageType {
		sl = wrapThisDescriptor(sl, desc, nil, file, i)
	}
	return sl
}

func wrapThisDescriptor(sl []*Descriptor, desc *DescriptorProto, parent *Descriptor, file *FileDescriptorProto, index int) []*Descriptor {
	sl = append(sl, newDescriptor(desc, parent, file, index))
	me := sl[len(sl)-1]
	for i, nested := range desc.NestedType {
		sl = wrapThisDescriptor(sl, nested, me, file, i)
	}
	return sl
}

// Construct the EnumDescriptor
func newEnumDescriptor(desc *EnumDescriptorProto, parent *Descriptor, file *FileDescriptorProto, index int) *EnumDescriptor {
	ed := &EnumDescriptor{
		common:              common{file},
		EnumDescriptorProto: desc,
		parent:              parent,
	}
	if parent == nil {
		ed.path = fmt.Sprintf("%d,%d", enumPath, index)
	} else {
		ed.path = fmt.Sprintf("%s,%d,%d", parent.path, messageEnumPath, index)
	}
	return ed
}

// Return a slice of all the EnumDescriptors defined within this file
func wrapEnumDescriptors(file *FileDescriptorProto, descs []*Descriptor) []*EnumDescriptor {
	sl := make([]*EnumDescriptor, 0, len(file.EnumType)+10)
	// Top-level enums.
	for i, enum := range file.EnumType {
		sl = append(sl, newEnumDescriptor(enum, nil, file, i))
	}

	return sl
}

// Return a slice of all the top-level ExtensionDescriptors defined within this file.
func wrapExtensions(file *FileDescriptorProto) []*FieldDescriptor {
	sl := make([]*FieldDescriptor, len(file.Extension))
	for i, field := range file.Extension {
		sl[i] = &FieldDescriptor{common{file}, field, nil}
	}
	return sl
}

// Construct the EnumDescriptor
func newServiceDescriptor(serv *ServiceDescriptorProto, file *FileDescriptorProto, index int) *ServiceDescriptor {
	sd := &ServiceDescriptor{
		common:                 common{file},
		ServiceDescriptorProto: serv,
	}
	sd.path = fmt.Sprintf("%d,%d", servicePath, index)
	return sd
}

// Return a slice of all the top-level ExtensionDescriptors defined within this file.
func wrapServiceDescriptors(file *FileDescriptorProto) []*ServiceDescriptor {
	sl := make([]*ServiceDescriptor, 0, len(file.Service)+10)
	for i, serve := range file.Service {
		sl = append(sl, newServiceDescriptor(serve, file, i))
	}
	return sl
}

func extractComments(file *FileDescriptor) {
	file.comments = make(map[string]*SourceCodeInfo_Location)
	for _, loc := range file.GetSourceCodeInfo().GetLocation() {
		if loc.LeadingComments == nil && loc.TrailingComments == nil {
			continue
		}
		//fmt.Println(loc.GoString())
		var p []string
		for _, n := range loc.Path {
			p = append(p, strconv.Itoa(int(n)))
		}
		key := strings.Join(p, ",")

		// Comment already exists
		if _, ok := file.comments[key]; ok {

			// While comment exists
			i := 1000
			_, ok2 := file.comments[key+","+fmt.Sprintf("%d", i)]
			for ok2 {
				_, ok2 = file.comments[key+","+fmt.Sprintf("%d", i)]
				i += 1000
			}
			// Assign a new comment
			file.comments[key+","+fmt.Sprintf("%d", i)] = loc
			//fmt.Println(key + "," + fmt.Sprintf("%d", i))
		} else {
			file.comments[key] = loc
			//fmt.Println(key)
		}

	}
}

// PrintComments prints any comments from the source .proto file.
// The path is a comma-separated list of integers.
// See descriptor.proto for its format.
func LeadingComments(path string, depth int) string {
	if loc, ok := currentFile.comments[path]; ok && loc.LeadingComments != nil {
		text := strings.TrimSuffix(loc.GetLeadingComments(), "\n")
		var s []string
		strCol := strings.Split(text, "\n")
		if len(strCol) == 1 {
			// Single line comments
			s = append(s, getIndentation(depth))
			s = append(s, "// ")
			s = append(s, strings.TrimSpace(strCol[0]))
			s = append(s, "\n")
		} else {
			// Multi-line comments
			if strings.Contains(text, "/*") || strings.Contains(text, "*/") {
				// Block comments cannot nest
				for _, line := range strCol {
					s = append(s, getIndentation(depth))
					s = append(s, "// ")
					s = append(s, strings.TrimSpace(line))
					s = append(s, "\n")
				}
			} else {
				s = append(s, getIndentation(depth))
				s = append(s, "/* ")
				s = append(s, strings.TrimSpace(strCol[0]))
				s = append(s, "\n ")
				for i := 1; i < len(strCol)-1; i += 1 {
					line := strCol[i]
					s = append(s, getIndentation(depth+1))
					s = append(s, strings.TrimSpace(line))
					s = append(s, "\n ")
				}
				s = append(s, getIndentation(depth+1))
				s = append(s, strings.TrimSpace(strCol[len(strCol)-1]))
				s = append(s, " */\n")
			}

		}
		return strings.Join(s, "")

	}

	return ""
}

func TrailingComments(path string, depth int) string {
	if loc, ok := currentFile.comments[path]; ok && loc.TrailingComments != nil {
		text := strings.TrimSuffix(loc.GetTrailingComments(), "\n")
		var s []string
		strCol := strings.Split(text, "\n")
		if len(strCol) == 1 {
			s = append(s, getIndentation(depth))
			s = append(s, "// ")
			s = append(s, strings.TrimSuffix(strings.TrimPrefix(strCol[0], " "), " "))
			s = append(s, "\n")
		} else {
			s = append(s, getIndentation(depth))
			s = append(s, "/* ")
			s = append(s, strings.TrimSuffix(strings.TrimPrefix(strCol[0], " "), " "))
			s = append(s, "\n ")
			for i := 1; i < len(strCol)-1; i += 1 {
				line := strCol[i]
				s = append(s, getIndentation(depth+1))
				s = append(s, strings.TrimSuffix(strings.TrimPrefix(line, " "), " "))
				s = append(s, "\n ")
			}
			s = append(s, getIndentation(depth+1))
			s = append(s, strings.TrimSuffix(strings.TrimPrefix(strCol[len(strCol)-1], " "), " "))
			s = append(s, " */\n")

		}
		return strings.Join(s, "")

	}

	return ""
}
