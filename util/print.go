package util

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	magenta = color.New(color.FgMagenta).SprintfFunc()
	blue    = color.New(color.FgBlue).SprintfFunc()
	red     = color.New(color.FgRed).SprintfFunc()
	green   = color.New(color.FgGreen).SprintfFunc()
)

func FormatAttributes(path string, attrs AttributeMap, print bool) {
	maxLength := 0
	for k, _ := range attrs {
		if len(k) > maxLength {
			maxLength = len(k)
		}
	}

	maxLength += 10

	lineFmt := fmt.Sprintf(" %%%ds %%s %%s\n", maxLength)

	dir, secretName := filepath.Split(path)
	var pathTokens []string
	if dir == "" {
		pathTokens = []string{"/"}
	} else {
		pathTokens = strings.Split(filepath.Clean(dir), "/")
	}

	fmt.Printf("Store » %s » %s\n", blue(strings.Join(pathTokens, " » ")), secretName)

	for k, v := range attrs {
		// Redact display of eyes-only attributes if -p is not set
		if v.EyesOnly {
			if print {
				v.Value = red(v.Value)
			} else {
				v.Value = red("<redacted>")
			}
		}
		if v.File {
			if !print {
				v.Value = green("<file content>")
			} else {
				v.Value = fmt.Sprintf("%s (use -w to write file to disk)", green("<file content>"))
			}
		}
		fmt.Printf(lineFmt, magenta(k), magenta("="), v.Value)
	}
}

func FormatDirectory(path string, level int) {
	dirPath := fmt.Sprintf("%s/%s", GetVaultPath(), path)

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
		if file.Name() == ".git" || file.Name() == "_vault.meta" {
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

func FormatKeyList(keys []MasterKey) {
	for idx, key := range keys {
		createdOn := time.Unix(int64(key.CreatedOn), 0)
		hash := md5.New()
		hash.Write([]byte(key.Data))

		fmt.Printf(" - #%d (%s) %s\n", idx, magenta(createdOn.Format("Tue, 02 Jan 2006, 15:04")), key.Comment)
		fmt.Printf("       %x\n", hash.Sum(nil))
	}
}
