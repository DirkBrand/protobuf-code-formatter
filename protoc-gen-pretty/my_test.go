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
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

var fileLocation string = "testdata/"

func TestAllOptions(t *testing.T) {
	fileName := "allOptionsTest.proto"
	parseAndTestFile(t, fileLocation+fileName)
}

func TestExtendWithComments(t *testing.T) {
	fileName := "extendWithCommentsTest.proto"
	parseAndTestFile(t, fileLocation+fileName)
}

func TestExtend(t *testing.T) {
	fileName := "extendTest.proto"
	parseAndTestFile(t, fileLocation+fileName)
}

func TestFieldOptions(t *testing.T) {
	fileName := "fieldOptionsTest.proto"
	parseAndTestFile(t, fileLocation+fileName)
}

func TestNestedMessage(t *testing.T) {
	fileName := "nestedMessageCommentsTest.proto"
	parseAndTestFile(t, fileLocation+fileName)
}

func TestOptionsCorrectlyFormatted(t *testing.T) {
	fileName := "optionCorrectlyFormattedTest.proto"
	parseAndTestFile(t, fileLocation+fileName)
}

func TestSampleComments(t *testing.T) {
	fileName := "sampleCommentsTest.proto"
	parseAndTestFile(t, fileLocation+fileName)
}

func TestServicesComments(t *testing.T) {
	fileName := "servicesCommentsTest.proto"
	parseAndTestFile(t, fileLocation+fileName)
}

func TestWalterTest1(t *testing.T) {
	fileName := "walterTest1.proto"
	parseAndTestFile(t, fileLocation+fileName)
}

func TestWalterTest2(t *testing.T) {
	fileName := "walterTest2.proto"
	parseAndTestFile(t, fileLocation+fileName)
}

func TestWalterTest3a(t *testing.T) {
	fileName := "walterTest3a.proto"
	parseAndTestFile(t, fileLocation+fileName)
}

func TestWalterTest3b(t *testing.T) {
	fileName := "walterTest3b.proto"
	parseAndTestFile(t, fileLocation+fileName)
}

func TestDescriptor(t *testing.T) {
	fileName := "descriptor.proto"
	parseAndTestFile(t, fileLocation+fileName)
}

func TestSample(t *testing.T) {
	fileName := "sample.proto"
	parseAndTestFile(t, fileLocation+fileName)
}

func TestDanglingComments(t *testing.T) {
	fileName := "commentsDangle.proto"
	parseAndTestFile(t, fileLocation+fileName)
}

func TestCommentedCode(t *testing.T) {
	fileName := "commentedCode.proto"
	parseAndTestFile(t, fileLocation+fileName)
}

// Negative Tests
func TestExtendCommentLimitation(t *testing.T) {
	fileName := "extendCommentsLimitationTest.proto"
	parseAndTestFile(t, fileLocation+fileName)
}

func TestOrderLostLimitation(t *testing.T) {
	fileName := "orderLostTest.proto"
	parseAndTestFile(t, fileLocation+fileName)
}

func TestUnattachedCommentsLostLimitation(t *testing.T) {
	fileName := "commentsStyleLostTest.proto"
	parseAndTestFile(t, fileLocation+fileName)
}

func parseAndTestFile(t *testing.T, filename string) {
	parser.FixFloatingComments(filename)

	d, err := parser.ParseFile(filename, "./")
	if err != nil {
		t.Error(err)
		os.Exit(1)
	} else {

		header := parser.ReadFileHeader(filename)

		formattedFile := d.Fmt(filename)
		formattedFile = strings.TrimSpace(formattedFile)
		if len(header) != 0 {
			formattedFile = header + formattedFile
		}

		// Test if formatted file can be parsed
		fo, err := os.Create("tempOutput.proto")
		if err != nil {
			t.Error(err)
		}
		defer os.Remove("tempOutput.proto")
		if len(formattedFile) > 0 {
			fo.WriteString(formattedFile)
		}
		fo.Close()

		_, err2 := parser.ParseFile("tempOutput.proto", "./", "../../../")
		if err2 != nil {
			t.Error(err2)
		}

		// Test if formatted string is equal to the Gold standard
		goldString, err := ioutil.ReadFile(strings.Split(filename, ".")[0] + "_Gold." + strings.Split(filename, ".")[1])
		if err != nil {
			t.Error(err)
		}

		if parser.Strcmp(formattedFile, strings.TrimSpace(string(goldString))) != 0 {
			t.Error("Failed the gold standard with: " + fmt.Sprintf("%v", parser.Strcmp(formattedFile, strings.TrimSpace(string(goldString)))))
		}

		return
	}
}
