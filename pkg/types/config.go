package types

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ConfigFile struct {
	Dir    string `yaml:"dir"`
	Server struct {
		Galaxy map[string]struct {
			URL string `yaml:"url"`
			Dir string `yaml:"dir"`
		} `yaml:"galaxy"`
		PYPI   map[string]string `yaml:"pypi"`
		Static map[string]string `yaml:"static"`
	} `yaml:"server"`
}

func (c *ConfigFile) Load(cfgFile string) {
	yamlFile, err := os.ReadFile(filepath.Clean(cfgFile))
	if err != nil {
		log.Printf("Read config file error: %v", err)
		os.Exit(1)
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal config file error: %v", err)
	}
}
