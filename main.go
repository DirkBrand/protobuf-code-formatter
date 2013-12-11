package main

import (
	proto "code.google.com/p/gogoprotobuf/proto"
	"fmt"
	parser "github.com/DirkBrand/protoc-gen-PBCF/parser"
	plugin "github.com/DirkBrand/protoc-gen-PBCF/plugin"
	"io/ioutil"
	"os"
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
		if len(Request.FileToGenerate) == 0 {
			Response.Error = proto.String("No files to generate")
		}

		formattedFiles := make(map[string]string, len(Request.GetFileToGenerate()))

		for _, fileToGen := range Request.GetFileToGenerate() {
			for _, protoFile := range Request.GetProtoFile() {
				if protoFile.GetName() == fileToGen {
					formattedFiles[fileToGen] = protoFile.FormattedGoString(0)
				}
			}
		}

		// Create the slice of response files
		Response.File = make([]*plugin.CodeGeneratorResponse_File, len(formattedFiles))

		i := 0
		for fileName, formatFile := range formattedFiles {

			fo, _ := os.Create("tempOutput.proto")
			//fmt.Print(formattedFile)
			fo.WriteString(formatFile)
			fo.Close()

			//fmt.Println(fileName)
			_, err2 := parser.ParseFile("tempOutput.proto", "./", "../../../")
			if err2 != nil {
				Response.Error = proto.String(err2.Error())
			} else {

				Response.File[i] = new(plugin.CodeGeneratorResponse_File)

				Response.File[i].Name = proto.String(`fixed_` + fileName)
				Response.File[i].Content = proto.String(formatFile)
				i += 1

			}
		}

		// Send back the results.
		data, err = proto.Marshal(Response)
		if err != nil {
			fmt.Println("failed to marshal output proto")
		}
		_, err = os.Stdout.Write(data)
		if err != nil {
			fmt.Println("failed to write output proto")
		}

	}
}
