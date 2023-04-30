package main

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/pieterclaerhout/export-komoot/komoot"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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
	return os.WriteFile(path, data, 0755)
}

func saveTourFile(data []byte, path string, tour komoot.Tour) error {
	if err := os.WriteFile(path, data, 0755); err != nil {
		return err
	}

	os.Chtimes(path, tour.Date, tour.ChangedAt)

	return nil
}
