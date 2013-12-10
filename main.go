package main

import (
	proto "code.google.com/p/gogoprotobuf/proto"
	"fmt"
	parser "github.com/DirkBrand/code-formatter/parser"
	plugin "github.com/DirkBrand/code-formatter/plugin"
	"io/ioutil"
	"os"
)

func main() {

	mainFile := "test.proto"

	data, err := ioutil.ReadAll(os.Stdin)

	if err != nil {
		fmt.Println(err)
	} else {
		fo, _ := os.Create("tempOutput.proto")

		// Declare the request and response structures
		Request := new(plugin.CodeGeneratorRequest)   // The input.
		Response := new(plugin.CodeGeneratorResponse) // The output.

		if err = proto.Unmarshal(data, Request); err != nil {
			fmt.Println(err)
		}
		if len(Request.FileToGenerate) == 0 {
			fmt.Println("No files to generate")
		}

		var formattedFile string
		for _, file := range Request.GetProtoFile() {
			if file.GetName() == mainFile {
				formattedFile = file.FormattedGoString(0)
			}
		}

		// Create the slice of response files
		Response.File = make([]*plugin.CodeGeneratorResponse_File, len(Request.GetFileToGenerate()))

		//fmt.Print(formattedFile)
		fo.WriteString(formattedFile)
		fo.Close()

		_, err2 := parser.ParseFile("tempOutput.proto", "./", "../../../")
		if err2 != nil {
			*Response.Error = fmt.Sprintf("%v", err2.Error())
		} else {

			Response.File[0] = new(plugin.CodeGeneratorResponse_File)

			Response.File[0].Name = proto.String(mainFile)
			Response.File[0].Content = proto.String(formattedFile)

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
}
