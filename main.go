package main

import (
	proto "code.google.com/p/gogoprotobuf/proto"
	descriptor "github.com/DirkBrand/protoc-gen-CF/descriptor"
	parser "github.com/DirkBrand/protoc-gen-CF/parser"
	plugin "github.com/DirkBrand/protoc-gen-CF/plugin"
	"io/ioutil"
	"os"
	//"path"
	//"strings"
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
					fileSet := descriptor.FileDescriptorSet{[]*descriptor.FileDescriptorProto{protoFile}, nil}
					formattedFiles[fileToGen] = fileSet.FormattedGoString(fileToGen)
				}
			}
		}

		// Create the slice of response files
		Response.File = make([]*plugin.CodeGeneratorResponse_File, len(formattedFiles))

		i := 0
		for fileName, formatFile := range formattedFiles {

			fo, _ := os.Create("tempOutput.proto")
			fo.WriteString(formatFile)
			fo.Close()

			_, err2 := parser.ParseFile("tempOutput.proto", "./", "../../../")
			os.Remove("tempOutput.proto")
			if err2 != nil {
				Response.Error = proto.String(err2.Error())
			} else {
				Response.File[i] = new(plugin.CodeGeneratorResponse_File)

				// Adds `_fixed`
				//fileName = path.Base(fileName)
				//fileName = strings.Split(fileName, ".")[0] + "_fixed." + strings.Split(fileName, ".")[1]

				Response.File[i].Name = proto.String(fileName)
				Response.File[i].Content = proto.String(formatFile)

				//os.Stderr.Write([]byte(Response.File[i].GetName() + "\n"))
				//os.Stderr.Write([]byte(Response.File[i].GetContent() + "\n"))
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
