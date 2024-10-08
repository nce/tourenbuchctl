package migrate

import (
	"bytes"
	"fmt"
	"os"

	"github.com/nce/tourenbuchctl/pkg/activity"
	"gopkg.in/yaml.v3"
)

func InsertOrUpdateVersion(filePath string, version string) (bool, error) {
	data, err := os.ReadFile(filePath + "/header.yaml")
	if err != nil {
		return false, fmt.Errorf("error reading file: %w", err)
	}

	var header activity.Header

	err = yaml.Unmarshal(data, &header)
	if err != nil {
		return false, fmt.Errorf("error unmarshalling YAML: %w", err)
	}

	header.Meta.Version = version

	// Marshal the updated data back to YAML
	var enc bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&enc)
	yamlEncoder.SetIndent(2)

	err = yamlEncoder.Encode(&header)
	if err != nil {
		return false, fmt.Errorf("error encoding YAML: %w", err)
	}

	//nolint: gosec
	err = os.WriteFile(filePath+"/header.yaml", enc.Bytes(), 0o644)
	if err != nil {
		return false, fmt.Errorf("error writing to file: %w", err)
	}

	return true, nil
}
