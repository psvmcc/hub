package types

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type GalaxyCollectionVersionInfoManifest struct {
	Format         int `json:"format"`
	CollectionInfo struct {
		Name          string            `json:"name"`
		Issues        string            `json:"issues"`
		Authors       []string          `json:"authors"`
		License       []string          `json:"license"`
		Version       string            `json:"version"`
		Homepage      string            `json:"homepage"`
		Namespace     string            `json:"namespace"`
		Repository    string            `json:"repository"`
		Description   string            `json:"description"`
		Dependencies  map[string]string `json:"dependencies"`
		Documentation string            `json:"documentation"`
		Tags          []string          `json:"tags"`
		Readme        string            `json:"readme"`
	} `json:"collection_info"`
	FileManifestFile struct {
		Name         string `json:"name"`
		Ftype        string `json:"ftype"`
		Format       int    `json:"format"`
		ChksumType   string `json:"chksum_type"`
		ChksumSha256 string `json:"chksum_sha256"`
	} `json:"file_manifest_file"`
}

type GalaxyCollectionVersionInfoFiles struct {
	Files []struct {
		Name         string `json:"name"`
		Ftype        string `json:"ftype"`
		Format       int    `json:"format"`
		ChksumType   any    `json:"chksum_type"`
		ChksumSha256 any    `json:"chksum_sha256"`
	} `json:"files"`
	Format int `json:"format"`
}

type GalaxyCollectionVersionInfo struct {
	Version         string    `json:"version"`
	Href            string    `json:"href"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	RequiresAnsible any       `json:"requires_ansible"`
	Marks           []any     `json:"marks"`
	Artifact        struct {
		Filename string `json:"filename"`
		Sha256   string `json:"sha256"`
		Size     int64  `json:"size"`
	} `json:"artifact"`
	Collection struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Href string `json:"href"`
	} `json:"collection"`
	DownloadURL string `json:"download_url"`
	Name        string `json:"name"`
	Namespace   struct {
		Name           string `json:"name"`
		MetadataSha256 string `json:"metadata_sha256"`
	} `json:"namespace"`
	Signatures interface{} `json:"signatures"`
	Metadata   struct {
		Authors       []string          `json:"authors"`
		Contents      []any             `json:"contents"`
		Dependencies  map[string]string `json:"dependencies"`
		Description   string            `json:"description"`
		Documentation string            `json:"documentation"`
		Homepage      string            `json:"homepage"`
		Issues        string            `json:"issues"`
		License       []string          `json:"license"`
		Repository    string            `json:"repository"`
		Tags          []string          `json:"tags"`
	} `json:"metadata"`
	GitURL       any                                 `json:"git_url"`
	GitCommitSha any                                 `json:"git_commit_sha"`
	Manifest     GalaxyCollectionVersionInfoManifest `json:"manifest"`
	Files        GalaxyCollectionVersionInfoFiles    `json:"files"`
}

func (c *GalaxyCollectionVersionInfo) ReadFromJSONFile(filePath string) error {
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
