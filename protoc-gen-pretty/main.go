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

package main

import (
	"io/ioutil"
	"os"

	descriptor "github.com/DirkBrand/protobuf-code-formatter/protoc-gen-pretty/descriptor"
	parser "github.com/DirkBrand/protobuf-code-formatter/protoc-gen-pretty/parser"
	plugin "github.com/DirkBrand/protobuf-code-formatter/protoc-gen-pretty/plugin"
	proto "github.com/gogo/protobuf/proto"
)

func main() {
	data, err := ioutil.ReadAll(os.Stdin)

	// Declare the request and response structures
	Request := new(plugin.CodeGeneratorRequest)   // The input.
	Response := new(plugin.CodeGeneratorResponse) // The output.

	if err != nil {
		Response.Error = proto.String("reading input")
	} else {
		if err = proto.Unmarshal(data, Request); err != nil {
			Response.Error = proto.String("parsing input proto")
		}
		if len(Request.GetFileToGenerate()) == 0 {
			Response.Error = proto.String("No files to generate")
		}

		formattedFiles := make(map[string]string)

		for _, fileToGen := range Request.GetFileToGenerate() {
			for _, protoFile := range Request.GetProtoFile() {
				if protoFile.GetName() == fileToGen {
					fileSet := descriptor.FileDescriptorSet{Request.GetProtoFile(), nil}
					formattedFiles[fileToGen] = fileSet.Fmt(fileToGen)

					if parser.CheckFloatingComments(fileToGen) {
						os.Stderr.WriteString("You had unattached comments that got lost.\n")
					}

					header := parser.ReadFileHeader(fileToGen)
					if len(header) != 0 {
						formattedFiles[fileToGen] = header + formattedFiles[fileToGen]
					}
					//os.Stderr.WriteString(fmt.Sprintf("%v", formattedFiles[fileToGen]))
				}
			}
		}

		// Create the slice of response files
		Response.File = make([]*plugin.CodeGeneratorResponse_File, len(formattedFiles))

		i := 0
		for fileName, formatFile := range formattedFiles {

			fo, err := os.Create("tempOutput.proto")
			if err != nil {
				panic(err)
			}
			defer os.Remove("tempOutput.proto")
			fo.WriteString(formatFile)
			fo.Close()

			_, err2 := parser.ParseFile("tempOutput.proto", "./", "../../../")
			if err2 != nil {
				Response.Error = proto.String(err2.Error())
			} else {
				Response.File[i] = new(plugin.CodeGeneratorResponse_File)

				Response.File[i].Name = proto.String(fileName)
				Response.File[i].Content = proto.String(formatFile)

				i += 1

			}
		}

		// Send back the results.
		data, err = proto.Marshal(Response)
		if err != nil {
			os.Stderr.Write([]byte("failed to marshal output proto"))
		}
		_, err = os.Stdout.Write(data)
		if err != nil {
			os.Stderr.Write([]byte("failed to write output proto"))
		}

	}
}
