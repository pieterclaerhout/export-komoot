package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
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

func formatJSON(data []byte) []byte {
	var out bytes.Buffer
	err := json.Indent(&out, data, "", "\t")
	if err != nil {
		return data
	}
	return out.Bytes()
}

func saveFormattedJSON(data []byte, path string) error {
	data = formatJSON(data)
	return ioutil.WriteFile(path, data, 0755)
}
