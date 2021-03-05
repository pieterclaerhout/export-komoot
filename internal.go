package main

import (
	"os"
	"path/filepath"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func deleteWithPattern(path string, pattern string) {
	items, err := filepath.Glob(filepath.Join(path, pattern))
	if err != nil {
		return
	}
	for _, item := range items {
		os.Remove(item)
	}
}
