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
	"flag"
	"fmt"
	parser "github.com/DirkBrand/protobuf-code-formatter/protoc-gen-pretty/parser"
	"os"
	"path/filepath"
	"strings"
)

var recurs *bool
var imp_path *string
var excluded []string

func main() {

	// FLAGS
	recurs = flag.Bool("r", false, "Indicates whether to recursively format the files in the argument folder.")
	imp_path = flag.String("proto_path", "./", "The path to find all relative imported .proto files.")
	exclude_dirs := flag.String("exclude_path", "None", "A list of directories that should not be included in the formatting (if done recursively)")

	flag.Parse()

	excluded = strings.Split(*exclude_dirs, ":")

	if len(os.Args) <= 1 || strings.HasPrefix(os.Args[len(os.Args)-1], "-") {
		fmt.Println(errors.New("Not enough arguments!"))
		os.Exit(1)
	}

	proto_path := os.Args[len(os.Args)-1]

	// Visit the directory / .proto file
	err := filepath.Walk(proto_path, fmtFn())
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("DONE")
	}

}

func fmtFn() filepath.WalkFunc {

	return func(pathThusFar string, f os.FileInfo, err error) error {

		if strings.HasSuffix(filepath.Dir(pathThusFar), f.Name()) {
			return nil
		}

		if f.IsDir() && !strings.HasSuffix(pathThusFar, string(os.PathSeparator)) {
			pathThusFar += string(os.PathSeparator)
		}
		if f.IsDir() {
			if *recurs && !stringInSlice(pathThusFar, excluded) {
				return nil
			} else {
				return filepath.SkipDir
			}
		} else if f.Mode().IsRegular() && strings.HasSuffix(f.Name(), ".proto") {
			parser.FixFloatingComments(pathThusFar)

			d, err := parser.ParseFile(pathThusFar, filepath.Dir(pathThusFar), *imp_path)
			if err != nil {
				fmt.Println("Parsing error in " + pathThusFar + "!")
				return err
			} else {
				header := parser.ReadFileHeader(pathThusFar)
				formattedFile := d.Fmt(f.Name())
				formattedFile = strings.TrimSpace(formattedFile)
				if len(header) != 0 {
					formattedFile = header + "\n" + formattedFile
				}

				fo, _ := os.Create(pathThusFar)

				fo.WriteString(formattedFile)
				fo.Close()

				fmt.Println("Successfully Formatted " + pathThusFar)
				return nil
			}
		} else {
			return fmt.Errorf("%v", errors.New(f.Name()+" cannot be processed."))
		}
	}

}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.HasPrefix(a, b) {
			return true
		}
	}
	return false
}
