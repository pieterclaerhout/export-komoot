package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pieterclaerhout/export-komoot/komoot"
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

func saveTourFile(data []byte, path string, tour komoot.Tour) error {

	err := ioutil.WriteFile(path, data, 0755)
	if err != nil {
		return err
	}

	os.Chtimes(path, tour.Date, tour.ChangedAt)

	return nil

}
