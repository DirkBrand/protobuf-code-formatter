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
	"strings"
)

func main() {

	// FLAGS
	r := flag.Bool("r", false, "Indicates whether to recursively format the files in the argument folder.")
	imp_path := flag.String("proto_path", "./", "The path to find all relative imported .proto files.")
	exclude_dirs := flag.String("exclude_dirs", "None", "A list of directories that should not be included in the formatting (if done recursively)")

	flag.Parse()

	excluded := strings.Split(*exclude_dirs, ":")

	if len(os.Args) <= 1 {
		panic(errors.New("Not enough arguments! You need atleast the .proto location. "))
	}

	proto_path := os.Args[len(os.Args)-1]

	d, err := os.Open(proto_path)
	if err != nil {
		fmt.Errorf("%v", err)
	}
	defer d.Close()
	fi, err := d.Readdir(-1)
	if err != nil {
		fmt.Errorf("%v", err)
	}
	for _, fi := range fi {
		// It is not a proper filename
		if !strings.HasSuffix(proto_path, string(os.PathSeparator)) {
			proto_path += string(os.PathSeparator)
		}
		// Visit the directory / .proto file
		visit(proto_path, *imp_path, excluded, fi, *r)
	}

}

func visit(pathThusFar string, imp_path string, exclude_paths []string, f os.FileInfo, recurs bool) {

	path := pathThusFar + f.Name()

	if f.IsDir() && !strings.HasSuffix(path, string(os.PathSeparator)) {
		path += string(os.PathSeparator)
	}

	if f.IsDir() && recurs && !stringInSlice(path, exclude_paths) {
		d, err := os.Open(path)
		if err != nil {
			fmt.Errorf("%v", err)
		}
		defer d.Close()
		fi, err := d.Readdir(-1)
		if err != nil {
			fmt.Errorf("%v", err)
		}
		for _, fi := range fi {
			visit(path, imp_path, exclude_paths, fi, recurs)
		}
	} else if f.Mode().IsRegular() && strings.HasSuffix(f.Name(), ".proto") {
		d, err := parser.ParseFile(path, pathThusFar, imp_path)
		if err != nil {
			panic(err)
		} else {
			fmt.Println("Formatted " + path)
			header := parser.ReadFileHeader(path)
			formattedFile := d.Fmt(f.Name())
			formattedFile = strings.TrimSpace(formattedFile)
			if len(header) != 0 {
				formattedFile = header + "\n" + formattedFile
			}

			fo, _ := os.Create(path)

			fo.WriteString(formattedFile)
			fo.Close()
		}
	} else {
		fmt.Errorf("%v", errors.New(f.Name()+" cannot be processed."))
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
