/*
Package config provides a simple configuration loader for the semantic-grep tool.
Contains a FindConfigFile function that searches for a configuration file
in a few standard locations, and a LoadConfig function that reads the
configuration file and returns a Config struct.
*/

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const DefaultConfigPath = "config.json"

type Config struct {
	ModelPath string `json:"model_path"`
}

func FindConfigFile() string {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Unable to determine current directory: %v\n", err)
		cwd = "."
	}

	locations := []string{
		filepath.Join(cwd, "config.json"),
		DefaultConfigPath,
		os.ExpandEnv("$HOME/.config/semantic-grep/config.json"),
		"/etc/semantic-grep/config.json",
	}

	for _, location := range locations {
		if _, err := os.Stat(location); err == nil {
			return location
		}
	}

	return ""
}

func LoadConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
