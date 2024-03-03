package types

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type GalaxyCollection struct {
	Href           string `json:"href"`
	Namespace      string `json:"namespace"`
	Name           string `json:"name"`
	Deprecated     bool   `json:"deprecated"`
	VersionsURL    string `json:"versions_url"`
	HighestVersion struct {
		Href    string `json:"href"`
		Version string `json:"version"`
	} `json:"highest_version"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DownloadCount int       `json:"download_count"`
}

func (c *GalaxyCollection) ReadFromJSONFile(filePath string) error {
	fileContent, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	err = json.Unmarshal(fileContent, c)
	if err != nil {
		return fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	return nil
}
