package migrate

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/nce/tourenbuchctl/pkg/activity"
	"gopkg.in/yaml.v3"
)

func ReduceElevationProfileToLabels(textDir string) (bool, error) {
	content, err := os.ReadFile("elevation.plt")
	if err != nil {
		return false, fmt.Errorf("error reading input file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	var filteredLines []string

	for _, line := range lines {
		if strings.HasPrefix(line, "set label") {
			filteredLines = append(filteredLines, line)
		}
	}

	//nolint: gosec
	err = os.WriteFile(textDir+"elevation.plt", []byte(strings.Join(filteredLines, "\n")), 0o644)
	if err != nil {
		return false, fmt.Errorf("error writing to input file: %w", err)
	}

	_, err = insertElevationProfileType(textDir)
	if err != nil {
		return false, fmt.Errorf("error writing to input file: %w", err)
	}

	return true, nil
}

func insertElevationProfileType(textDir string) (bool, error) {
	data, err := os.ReadFile(textDir + "header.yaml")
	if err != nil {
		return false, fmt.Errorf("error reading file: %w", err)
	}

	var header activity.Header

	err = yaml.Unmarshal(data, &header)
	if err != nil {
		return false, fmt.Errorf("error unmarshalling YAML: %w", err)
	}

	var layout string

	activityType, err := activity.GetFromHeader[string](textDir, "Activity.Type")
	if err != nil {
		return false, fmt.Errorf("error getting activity type: %w", err)
	}

	switch activityType {
	case "skitour":
		layout = "right-axis-filtered-layout"
	case "wandern":
		layout = "right-axis-filtered-layout"
	case "mtb":
		layout = "left-axis-layout"
	}

	// Check if elevationProfile key exists under layout
	if header.Layout.ElevationProfileType == "" {
		// If elevationProfile key does not exist, add it
		header.Layout.ElevationProfileType = layout
	}
	// Marshal the updated data back to YAML
	var enc bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&enc)
	yamlEncoder.SetIndent(2)

	err = yamlEncoder.Encode(&header)
	if err != nil {
		return false, fmt.Errorf("error encoding YAML: %w", err)
	}

	//nolint: gosec
	err = os.WriteFile(textDir+"header.yaml", enc.Bytes(), 0o644)
	if err != nil {
		return false, fmt.Errorf("error writing to file: %w", err)
	}

	return true, nil
}
