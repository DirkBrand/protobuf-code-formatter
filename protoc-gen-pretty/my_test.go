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
		os.Remove("tempOutput.proto")
		if err2 != nil {
			fmt.Println(err2)
			t.Fail()
		} else {
			fmt.Println(filename + " CORRECTLY PARSED")
		}
	}
}
