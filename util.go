package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

type Secret struct {
	Salt     string   `json:"salt"`
	Nonce    string   `json:"nonce"`
	Data     string   `json:"data"`
	EyesOnly []string `json:"eyesonly"`
}

var (
	magenta = color.New(color.FgMagenta).SprintfFunc()
	blue    = color.New(color.FgBlue).SprintfFunc()
	red     = color.New(color.FgRed).SprintfFunc()
)

func FormatAttributes(path string, attrs map[string]string, eyesOnly []string, print bool) {
	maxLength := 0
	for k, _ := range attrs {
		if len(k) > maxLength {
			maxLength = len(k)
		}
	}

	maxLength += 10

	lineFmt := fmt.Sprintf(" %%%ds %%s %%s\n", maxLength)

	dir, secretName := filepath.Split(path)
	pathTokens := strings.Split(filepath.Clean(dir), "/")
	// pathTokens := strings.Split(path, "/")
	// secretName := pathTokens[len(pathTokens)-1]
	// pathTokens = pathTokens[:len(pathTokens)-1]

	fmt.Printf("Store » %s » %s\n", blue(strings.Join(pathTokens, " » ")), secretName)

	for k, v := range attrs {
		// Redact display of eyes-only attributes if -p is not set
		if StringArrayContains(eyesOnly, k) {
			if print {
				v = red(v)
			} else {
				v = red("<redacted>")
			}
		}
		fmt.Printf(lineFmt, magenta(k), magenta("="), v)
	}
}

func FormatDirectory(path string, level int) {
	dirPath := fmt.Sprintf("%s/%s", vaultDir, path)

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return
	}

	// Display styled representation of current directory on first line
	if level == 0 {
		var pathTokens []string
		if path == "/" {
			pathTokens = []string{"/"}
		} else {
			pathTokens = strings.Split(filepath.Clean(path), "/")
		}

		fmt.Printf("Store » %s\n", blue(strings.Join(pathTokens, " » ")))
	}

	indent := ""
	for i := 0; i < level*2; i++ {
		indent = fmt.Sprintf("%s ", indent)
	}

	for _, file := range files {
		if file.Name() == ".git" {
			continue
		}

		if file.IsDir() {
			fmt.Printf("%s  » %s\n", indent, blue(file.Name()))

			FormatDirectory(fmt.Sprintf("%s/%s", path, file.Name()), level+1)
		} else {
			fmt.Printf("%s  - %s\n", indent, file.Name())
		}
	}
}

func StringArrayContains(arr []string, item string) bool {
	for _, v := range arr {
		if v == item {
			return true
		}
	}
	return false
}

func RemoveFromSlice(arr []string, item string) []string {
	newArr := make([]string, 0)
	for _, v := range arr {
		if v != item {
			newArr = append(newArr, v)
		}
	}
	return newArr
}
