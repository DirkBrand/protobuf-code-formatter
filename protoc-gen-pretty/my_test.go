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
	"errors"
	"fmt"
	parser "github.com/DirkBrand/protobuf-code-formatter/protoc-gen-pretty/parser"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

var fileLocation string = "testdata/"

func TestAllOptions(t *testing.T) {
	fileName := "allOptionsTest.proto"

	if res, err := parseAndTestFile(fileLocation + fileName); !res {
		t.Errorf("%v", err)
	} else {
		fmt.Println(fileName + " <TEST PASSED>")
	}
}

func TestExtendWithComments(t *testing.T) {
	fileName := "extendWithCommentsTest.proto"

	if res, err := parseAndTestFile(fileLocation + fileName); !res {
		t.Errorf("%v", err)
	} else {
		fmt.Println(fileName + " <TEST PASSED>")
	}
}

func TestExtend(t *testing.T) {
	fileName := "extendTest.proto"

	if res, err := parseAndTestFile(fileLocation + fileName); !res {
		t.Errorf("%v", err)
	} else {
		fmt.Println(fileName + " <TEST PASSED>")
	}
}

func TestFieldOptions(t *testing.T) {
	fileName := "fieldOptionsTest.proto"
	res, err := parseAndTestFile(fileLocation + fileName)
	if !res {
		t.Errorf("%v", err)
	} else {
		fmt.Println(fileName + " <TEST PASSED>")
	}
}

func TestNestedMessage(t *testing.T) {
	fileName := "nestedMessageCommentsTest.proto"
	res, err := parseAndTestFile(fileLocation + fileName)
	if !res {
		t.Errorf("%v", err)
	} else {
		fmt.Println(fileName + " <TEST PASSED>")
	}
}

func TestOptionsCorrectlyFormatted(t *testing.T) {
	fileName := "optionCorrectlyFormattedTest.proto"
	res, err := parseAndTestFile(fileLocation + fileName)
	if !res {
		t.Errorf("%v", err)
	} else {
		fmt.Println(fileName + " <TEST PASSED>")
	}
}

func TestSampleComments(t *testing.T) {
	fileName := "sampleCommentsTest.proto"
	res, err := parseAndTestFile(fileLocation + fileName)
	if !res {
		t.Errorf("%v", err)
	} else {
		fmt.Println(fileName + " <TEST PASSED>")
	}
}

func TestServicesComments(t *testing.T) {
	fileName := "servicesCommentsTest.proto"
	res, err := parseAndTestFile(fileLocation + fileName)
	if !res {
		t.Errorf("%v", err)
	} else {
		fmt.Println(fileName + " <TEST PASSED>")
	}
}

func TestWalterTest1(t *testing.T) {
	fileName := "walterTest1.proto"
	res, err := parseAndTestFile(fileLocation + fileName)
	if !res {
		t.Errorf("%v", err)
	} else {
		fmt.Println(fileName + " <TEST PASSED>")
	}
}

// Negative Tests
func TestExtendCommentLimitation(t *testing.T) {
	fileName := "extendCommentsLimitationTest.proto"
	res, err := parseAndTestFile(fileLocation + fileName)
	if !res {
		t.Errorf("%v", err)
	} else {
		fmt.Println(fileName + " <TEST PASSED>")
	}
}

func TestOrderLostLimitation(t *testing.T) {
	fileName := "orderLostTest.proto"
	res, err := parseAndTestFile(fileLocation + fileName)
	if !res {
		t.Errorf("%v", err)
	} else {
		fmt.Println(fileName + " <TEST PASSED>")
	}
}

func TestUnattachedCommentsLostLimitation(t *testing.T) {
	fileName := "commentsStyleLostTest.proto"
	res, err := parseAndTestFile(fileLocation + fileName)
	if !res {
		t.Errorf("%v", err)
	} else {
		fmt.Println(fileName + " <TEST PASSED>")
	}
}

func parseAndTestFile(filename string) (bool, error) {
	d, err := parser.ParseFile(filename, "./")
	if err != nil {
		return false, err
	} else {

		header := parser.ReadFileHeader(filename)

		formattedFile := d.Fmt(filename)
		formattedFile = strings.TrimSpace(formattedFile)
		if len(header) != 0 {
			formattedFile = header + "\n" + formattedFile
		}

		// Test if formatted file can be parsed
		fo, err := os.Create("tempOutput.proto")
		if err != nil {
			panic(err)
		}
		if len(formattedFile) > 0 {
			fo.WriteString(formattedFile)
		}
		fo.Close()

		_, err2 := parser.ParseFile("tempOutput.proto", "./", "../../../")
		defer os.Remove("tempOutput.proto")
		if err2 != nil {
			return false, err2
		}

		// Test if formatted string is equal to the Gold standard
		goldString, err := ioutil.ReadFile(strings.Split(filename, ".")[0] + "_Gold." + strings.Split(filename, ".")[1])
		if err != nil {
			panic(err)
		}

		if parser.Strcmp(formattedFile, strings.TrimSpace(string(goldString))) != 0 {
			fmt.Println("Failed: " + filename)
			os.Exit(-1)
			return false, errors.New("Failed the gold standard with: " + fmt.Sprintf("%v", parser.Strcmp(formattedFile, strings.TrimSpace(string(goldString)))))
		}

		return true, nil

	}
}
