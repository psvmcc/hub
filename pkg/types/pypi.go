package types

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type PypiMetadata struct {
	Files []struct {
		CoreMetadata         interface{} `json:"core-metadata"`
		DataDistInfoMetadata interface{} `json:"data-dist-info-metadata"`
		Filename             string      `json:"filename"`
		Hashes               struct {
			Sha256 string `json:"sha256"`
		} `json:"hashes"`
		RequiresPython string      `json:"requires-python"`
		Size           int         `json:"size"`
		UploadTime     time.Time   `json:"upload-time"`
		URL            string      `json:"url"`
		Yanked         interface{} `json:"yanked"`
	} `json:"files"`
	Meta struct {
		LastSerial int    `json:"_last-serial"`
		APIVersion string `json:"api-version"`
	} `json:"meta"`
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
}

func (p *PypiMetadata) ReadFromJSONFile(filePath string) error {
	fileContent, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	err = json.Unmarshal(fileContent, p)
	if err != nil {
		return fmt.Errorf("error unmarshalling JSON: %v", err)
	}
	return nil
}
