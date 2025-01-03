package wadm

import (
	"encoding/json"
	"os"

	yaml "github.com/goccy/go-yaml"
)

func ParseManifest(data []byte) (*Manifest, error) {
	manifest, err := ParseYAMLManifest(data)
	if err == nil {
		return manifest, nil
	}

	return ParseJSONManifest(data)
}

func ParseJSONManifest(data []byte) (*Manifest, error) {
	manifest := &Manifest{}
	return manifest, json.Unmarshal(data, manifest)
}

func ParseYAMLManifest(data []byte) (*Manifest, error) {
	manifest := &Manifest{}
	return manifest, yaml.Unmarshal(data, manifest)
}

func LoadManifest(path string) (*Manifest, error) {
	rawManifest, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseManifest(rawManifest)
}
