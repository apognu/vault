package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/apognu/vault/crypt"
	"github.com/apognu/vault/util"
	"github.com/atotto/clipboard"
)

func listSecrets(path string) {
	dirPath := fmt.Sprintf("%s/%s", util.GetVaultPath(), path)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		logrus.Fatal("secret does not exist")
	}

	util.FormatDirectory(path, 0)
}

func showSecret(path string, print bool, clip bool, clipAttr string, write bool, writeFiles []string) {
	_, attrs := crypt.GetSecret(path)

	if clipAttr == "" {
		if attrs.EyesOnlyCount() == 1 {
			clipAttr = attrs.FindFirstEyesOnly()
		} else {
			clipAttr = "password"
		}
	}

	if clip {
		if attrs[clipAttr] != nil {
			clipboard.WriteAll(attrs[clipAttr].Value)
			logrus.Infof("attribute '%s' of '%s' was copied to your clipboard", clipAttr, path)
			return
		} else {
			logrus.Fatalf("could not read attribute '%s'", clipAttr)
		}
	}

	if write {
		WriteFiles(path, attrs, writeFiles)
		return
	}

	util.FormatAttributes(path, attrs, print)
}

func addSecret(path string, attributes map[string]string, generatorLength int, edit bool, editedAttrs []string) {
	// Check if the secret already exists in ADD mode
	filePath := fmt.Sprintf("%s/%s", util.GetVaultPath(), path)
	if !edit {
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			logrus.Fatal("secret already exists")
		}
	}

	attrs := make(util.AttributeMap)
	for k, v := range attributes {
		attrs[k] = &util.Attribute{
			Value: v,
		}
	}

	crypt.SetSecret(path, attrs, generatorLength, edit, editedAttrs, false)
}

func editSecret(path string, newAttrs map[string]string, deletedAttrs []string, generatorLength int) {
	_, attrs := crypt.GetSecret(path)
	editedAttrs := make([]string, 0)

	// Replace old attributes with new ones
	for k, v := range newAttrs {
		if attrs[k] == nil {
			attrs[k] = &util.Attribute{Value: v}
		} else {
			attrs[k].Value = v
		}
		editedAttrs = append(editedAttrs, k)
	}

	// Remove deleted attributes from the map
	for _, k := range deletedAttrs {
		delete(attrs, k)
	}

	crypt.SetSecret(path, attrs, generatorLength, true, editedAttrs, false)
}

func deleteSecret(path string) {
	err := os.Remove(fmt.Sprintf("%s/%s", util.GetVaultPath(), path))
	if err != nil {
		logrus.Fatalf("could not remove secret: %s", err)
	}

	logrus.Infof("secret '%s' deleted successfully", path)
	util.GitCommit(path, util.GIT_DELETE, "")

	// Remove any empty parent directory
	for {
		dir, _ := filepath.Split(filepath.Clean(path))
		if dir == "" {
			break
		}
		err := os.Remove(fmt.Sprintf("%s/%s", util.GetVaultPath(), dir))
		if err != nil {
			return
		}
		path = dir
	}
}

func WriteFiles(path string, attrs util.AttributeMap, writeFiles []string) {
	fileAttrs := make(util.AttributeMap)
	for n, a := range attrs {
		if a.File {
			if len(writeFiles) != 0 {
				if !util.StringArrayContains(writeFiles, n) {
					continue
				}
			}
			fileAttrs[n] = a
		}
	}

	if len(fileAttrs) == 0 {
		logrus.Fatal("no file attribute matching what you requested")
	}

	dirs, secretName := filepath.Split(path)
	dir := ""
	if len(dirs) == 0 {
		dir = fmt.Sprintf("vault-%s", secretName)
	} else {
		dir = fmt.Sprintf("vault-%s-%s", strings.Join(strings.Split(filepath.Clean(dirs), string(os.PathSeparator)), "-"), secretName)
	}

	if err := os.Mkdir(dir, 0700); err != nil {
		logrus.Fatalf("could not create directory %s", err)
	}

	for n, a := range fileAttrs {
		fileName := fmt.Sprintf("%s/%s", dir, n)
		file, err := os.Create(fileName)
		if err != nil {
			logrus.Fatalf("could not create output file: %s", err)
		}
		defer file.Close()
		file.Chmod(0400)
		file.Write([]byte(a.Value))

		logrus.Infof("attribute written to '%s'", fileName)
	}
}
