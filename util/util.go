package util

import (
	"os"
	"path/filepath"
	"strings"
)

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

func ShouldFileBeWalked(path string) (bool, error) {
	if strings.HasSuffix(path, ".git") {
		return false, filepath.SkipDir
	}
	if strings.HasSuffix(path, "_vault.meta") || strings.HasSuffix(path, "_vault.meta.new") {
		return false, nil
	}
	if f, _ := os.Stat(path); f.IsDir() {
		return false, nil
	}
	return true, nil
}
