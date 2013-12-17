package main

import (
	"fmt"
	parser "github.com/DirkBrand/protobuf-code-formatter/protoc-gen-pretty/parser"
	"os"
	"strings"
	"testing"
	//descriptor "github.com/DirkBrand/test/descriptor"
)

func TestAllFiles(t *testing.T) {
	fileLocation := "testdata/"
	d, err := os.Open(fileLocation)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fi, err := d.Readdir(-1)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, f := range fi {
		if strings.HasSuffix(f.Name(), ".proto") {
			fmt.Println("Testing " + f.Name())
			parseAndTestFile(t, fileLocation+f.Name())
		}
	}

}

func parseAndTestFile(t *testing.T, filename string) {

	d, err := parser.ParseFile(filename, "./")
	if err != nil {
		t.Errorf("%v", err)
	} else {
		fo, _ := os.Create("tempOutput.proto")

		var formattedFile string
		formattedFile = d.FormattedGoString(filename)

		//fmt.Print(formattedFile)

		fo.WriteString(formattedFile)
		fo.Close()

		_, err2 := parser.ParseFile("tempOutput.proto", "./", "../../../")
		//os.Remove("tempOutput.proto")
		if err2 != nil {
			fmt.Println(err2)
			t.Fail()
		} else {
			fmt.Println(filename + " CORRECTLY PARSED")
		}
	}
}
