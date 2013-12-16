package test

import (
	"fmt"
	parser "github.com/DirkBrand/test/parser"
	"os"
	"testing"
	//descriptor "github.com/DirkBrand/test/descriptor"
)

func TestFileParsing(t *testing.T) {
	testFile := "../test_protocol_buffers/group_comments.proto"

	d, err := parser.ParseFile(testFile, "./", "../../../")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	} else {
		fo, _ := os.Create("tempOutput.proto")

		formattedFile := d.FormattedGoString(testFile)

		fmt.Print(formattedFile)

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
