package parser

import (
	"bufio"
	"io"
	"os"
	"strings"
)

func ReadFileHeader(filename string) string {
	var s string

	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	r := bufio.NewReader(f)
	gap := 0
	for {
		path, err := r.ReadString(10) // 0x0A separator = newline
		if err == io.EOF {
			// do something here
			break
		} else if err != nil {
			panic(err)
		}

		path = strings.TrimSpace(path)

		if strings.HasPrefix(path, "//") {
			gap = 0
			s += path + "\n"
		} else if len(path) == 0 {
			if gap == 0 {
				s += "\n"
			}
			gap += 1
		} else {
			if gap == 0 {
				s = ""
				break
			} else {
				break
			}
		}
	}

	return s
}

func Strcmp(a, b string) int {
	var min = len(b)
	if len(a) < len(b) {
		min = len(a)
	}
	var diff int
	for i := 0; i < min && diff == 0; i++ {
		diff = int(a[i]) - int(b[i])
	}
	if diff == 0 {
		diff = len(a) - len(b)
	}
	return diff
}
