package main

import (
	"fmt"
	parser "github.com/DirkBrand/test/parser"
	"os"
	"testing"
	//descriptor "github.com/DirkBrand/test/descriptor"
)

func TestFileParsing(t *testing.T) {
	testFile := "testdata/walter_test1.proto"

	d, err := parser.ParseFile(testFile, "./")
	if err != nil {
		t.Errorf("%v", err)
	} else {
		fo, _ := os.Create("tempOutput.proto")

		var formattedFile string
		formattedFile = d.FormattedGoString(testFile)

		//fmt.Print(formattedFile)

		fo.WriteString(formattedFile)
		fo.Close()

		_, err2 := parser.ParseFile("tempOutput.proto", "./", "../../../")
		os.Remove("tempOutput.proto")
		if err2 != nil {
			fmt.Println(err2)
			t.Fail()
		} else {
			fmt.Println("CORRECTLY PARSED")
		}
	}

}
