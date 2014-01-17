package parser

import (
	"bufio"
	//"fmt"
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

func CheckFloatingComments(filename string) bool {
	var file []string
	changed := false

	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := bufio.NewReader(f)

	// Generate slice of program
	for {
		path, err := r.ReadString(10) // 0x0A separator = newline
		if err == io.EOF {
			file = append(file, path)
			break
		} else if err != nil {
			panic(err)
		}
		file = append(file, path)
	}

	for i, _ := range file {
		// Start line comment
		if strings.HasPrefix(strings.TrimSpace(file[i]), "//") {
			if (i > 0 && len(strings.TrimSpace(file[i-1])) == 0) || i == 0 {

				for strings.HasPrefix(strings.TrimSpace(file[i]), "//") && i < len(file) {
					i += 1
				}

				// Dangling comment found
				if i < len(file) && len(strings.TrimSpace(file[i])) == 0 {
					for i < len(file) && len(strings.TrimSpace(file[i])) == 0 {
						return true
					}
				}
			}
		}

		// Start block comment
		if strings.HasPrefix(file[i], "/*") {
			if (i > 0 && len(strings.TrimSpace(file[i-1])) == 0) || i == 0 {
				for i < len(file) {
					if strings.Contains(file[i], "*/") {
						i += 1
						break
					}
					i += 1
				}
				// Dangling comment found
				if i < len(file) && len(strings.TrimSpace(file[i])) == 0 {
					for i < len(file) && len(strings.TrimSpace(file[i])) == 0 {
						return true
					}
				}
			}
		}

		if i == len(file)-1 {
			break
		}

	}

	return changed

}

func FixFloatingComments(filename string) {
	var file []string

	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := bufio.NewReader(f)

	// Generate slice of program
	for {
		path, err := r.ReadString(10) // 0x0A separator = newline
		if err == io.EOF {
			file = append(file, path)
			break
		} else if err != nil {
			panic(err)
		}
		file = append(file, path)
	}

	for i, _ := range file {
		// Start line comment
		if strings.HasPrefix(strings.TrimSpace(file[i]), "//") {
			if (i > 0 && len(strings.TrimSpace(file[i-1])) == 0) || i == 0 {

				for strings.HasPrefix(strings.TrimSpace(file[i]), "//") && i < len(file) {
					i += 1
				}

				// Dangling comment found
				if i < len(file) && len(strings.TrimSpace(file[i])) == 0 {
					for i < len(file) && len(strings.TrimSpace(file[i])) == 0 {
						file[i] = "//\n"
						i += 1
					}
				}
			}
		}

		// Start block comment
		if strings.HasPrefix(file[i], "/*") {
			if (i > 0 && len(strings.TrimSpace(file[i-1])) == 0) || i == 0 {
				file[i] = strings.Replace(file[i], "/*", "", -1)
				for i < len(file) {
					if strings.Contains(file[i], "*/") {
						file[i] = strings.Replace(file[i], "*/", "", -1)
						file[i] = "// " + file[i]
						i += 1
						break
					}
					file[i] = "// " + file[i]
					file[i] = strings.Replace(file[i], "*", "", -1)
					i += 1
				}
				// Dangling comment found
				if i < len(file) && len(strings.TrimSpace(file[i])) == 0 {
					for i < len(file) && len(strings.TrimSpace(file[i])) == 0 {
						file = append(file[:i], file[i+1:]...)
					}
				}
			}
		}

		if i == len(file)-1 {
			break
		}

	}

	// Overwrite file

	fo, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	fo.WriteString(strings.Join(file, ""))
	fo.Close()

	//fmt.Println(strings.Join(file, ""))
}
