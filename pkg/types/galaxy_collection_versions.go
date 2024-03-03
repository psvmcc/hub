package types

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type GalaxyCollectionVersion struct {
	Version         string    `json:"version"`
	Href            string    `json:"href"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	RequiresAnsible string    `json:"requires_ansible"`
	Marks           []any     `json:"marks"`
}

type GalaxyCollectionVersions struct {
	Meta struct {
		Count int `json:"count"`
	} `json:"meta"`
	Links struct {
		First    string `json:"first"`
		Previous string `json:"previous"`
		Next     string `json:"next"`
		Last     string `json:"last"`
	} `json:"links"`
	Data []GalaxyCollectionVersion `json:"data"`
}

func (c *GalaxyCollectionVersions) ReadFromJSONFile(filePath, key, namespace, name string) error {
	fileContent, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	err = json.Unmarshal(fileContent, c)
	if err != nil {
		return fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	first, err := url.Parse(c.Links.First)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
	}
	c.Links.First = fmt.Sprintf("/galaxy/%s/api/v3/collections/%s/%s/versions/?%s", key, namespace, name, first.RawQuery)

	last, err := url.Parse(c.Links.Last)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
	}
	c.Links.Last = fmt.Sprintf("/galaxy/%s/api/v3/collections/%s/%s/versions/?%s", key, namespace, name, last.RawQuery)

	if c.Links.Next != "" {
		next, err := url.Parse(c.Links.Next)
		if err != nil {
			fmt.Println("Error parsing URL:", err)
		}
		c.Links.Next = fmt.Sprintf("/galaxy/%s/api/v3/collections/%s/%s/versions/?%s", key, namespace, name, next.RawQuery)
	}
	if c.Links.Previous != "" {
		previous, err := url.Parse(c.Links.Previous)
		if err != nil {
			fmt.Println("Error parsing URL:", err)
		}
		c.Links.Previous = fmt.Sprintf("/galaxy/%s/api/v3/collections/%s/%s/versions/?%s", key, namespace, name, previous.RawQuery)
	}

	for i := range c.Data {
		c.Data[i].Href = fmt.Sprintf("/galaxy/%s/api/v3/collections/%s/%s/versions/%s/", key, namespace, name, c.Data[i].Version)
	}

	return nil
}
