package activity

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var (
	ErrFieldNotFound = errors.New("field not found in struct")
	ErrTypeNotAssert = errors.New("could not assert type in struct")
)

type Header struct {
	Meta struct {
		Version string `yaml:"version,omitempty"`
	} `yaml:"meta"`
	Activity struct {
		Wandern       bool   `yaml:"wandern,omitempty"`
		Skitour       bool   `yaml:"skitour,omitempty"`
		MTB           bool   `yaml:"mtb,omitempty"`
		Type          string `yaml:"type"`
		Date          string `yaml:"date"`
		Title         string `yaml:"title"`
		PointOfOrigin struct {
			Name   string `yaml:"name"`
			Qr     string `yaml:"qr"`
			Region string `yaml:"region"`
		} `yaml:"pointOfOrigin"`
		Season       string `yaml:"season"`
		Rating       string `yaml:"rating"`
		Company      string `yaml:"company"`
		Restaurant   string `yaml:"restaurant"`
		Difficulty   string `yaml:"difficulty,omitempty"`
		LLB          string `yaml:"llb,omitempty"`
		MaxElevation string `yaml:"maxElevation"`
	} `yaml:"activity"`
	Layout struct {
		HeadElevationProfile        bool    `yaml:"headElevationProfile"`
		ElevationProfileType        string  `yaml:"elevationProfileType,omitempty"`
		ElevationProfileRightMargin float32 `yaml:"elevationProfileRightMargin"`
		TableSize                   float32 `yaml:"tableSize"`
		MapSize                     float32 `yaml:"mapSize"`
		MapHeight                   int     `yaml:"mapHeight"`
		Linespread                  float32 `yaml:"linespread"`
	} `yaml:"layout"`
	Stats struct {
		Ascent      string `yaml:"ascent"`
		Distance    string `yaml:"distance"`
		MovingTime  string `yaml:"movingTime"`
		OverallTime string `yaml:"overallTime"`
		StartTime   string `yaml:"startTime"`
		SummitTime  string `yaml:"summitTime"`
		Puls        string `yaml:"puls,omitempty"`
	} `yaml:"stats"`
}

func GetFromHeader[T any](dir string, fields ...string) (map[string]T, error) {
	data, err := os.ReadFile(dir + "/header.yaml")
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var act Header

	err = yaml.Unmarshal(data, &act)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML: %w", err)
	}

	results := make(map[string]T, len(fields))

	for _, field := range fields {
		value, err := searchField[T](&act, field)
		if err != nil {
			return nil, fmt.Errorf("error searching field: %w", err)
		}

		results[field] = value
	}

	return results, nil
}

func searchField[T any](v interface{}, path string) (T, error) { //nolint:ireturn
	keys := strings.Split(path, ".")
	header := reflect.ValueOf(v)

	// Traverse the struct hierarchy using the keys
	for _, key := range keys {
		// Check if we are dealing with a pointer and dereference it
		if header.Kind() == reflect.Ptr {
			header = header.Elem()
		}

		// Get the field by name
		header = header.FieldByName(key)
		if !header.IsValid() {
			var zero T

			return zero, fmt.Errorf("field %s not found: %w", key, ErrFieldNotFound)
		}
	}

	// Type assert the value to the desired type T
	value, ok := header.Interface().(T)
	if !ok {
		var zero T

		return zero, fmt.Errorf("field %s cannot be asserted to the expected type: %w", path, ErrTypeNotAssert)
	}

	return value, nil
}

// Updates the existing stats-yaml structure with new data.
//
//gocyclo:ignore
func (a *Activity) updateActivity(dir string) error {
	file := dir + "header.yaml"

	yamlFile, err := os.ReadFile(file)
	if err != nil {
		log.Error().Str("filename", file).Msg("Error reading file")

		return fmt.Errorf("reading file %w", err)
	}

	// Parse the YAML file into a node tree
	var root yaml.Node

	err = yaml.Unmarshal(yamlFile, &root)
	if err != nil {
		log.Error().Str("filename", file).Msg("Error unmarshalling file")

		return fmt.Errorf("unmarshalling file %w", err)
	}

	// No idea. This is written by AI
	// This modifies just one nested Yaml Node, without touching/killing the
	// rest of the file. It updates the statistic data (distance/ascent) of the activity

	// Navigate to the "stats" node and modify it
	// Traverse the document to find the "stats" key
	for i := range len(root.Content) {
		if root.Content[i].Kind == yaml.MappingNode {
			for j := 0; j < len(root.Content[i].Content); j += 2 {
				keyNode := root.Content[i].Content[j]
				valueNode := root.Content[i].Content[j+1]

				if keyNode.Value == "stats" {
					// Modify the stats node
					// Example: Modify a specific key-value pair within the "stats" node
					for k := 0; k < len(valueNode.Content); k += 2 {
						keyNode := valueNode.Content[k]
						value := valueNode.Content[k+1]

						switch keyNode.Value {
						case "ascent":
							value.Value = normalizeElevation(a.Tb.Ascent)
						case "distance":
							value.Value = a.normalizeDistance()
						case "movingTime":
							value.Value = normalizeDuration(a.Tb.MovingTime)
						case "overallTime":
							value.Value = normalizeDuration(a.Tb.ElapsedTime)
						case "startTime":
							value.Value = a.normalizeStartTime()
						}

						value.Tag = "!!str"
					}
				}
			}
		}
	}

	// Serialize the modified node tree back to a YAML string
	output, err := yaml.Marshal(&root)
	if err != nil {
		log.Error().Str("filename", file).Msg("Serializing to yaml failed")

		return fmt.Errorf("serialzing yaml %w", err)
	}

	// Write the modified YAML string back to the file
	//nolint: gosec
	err = os.WriteFile(file, output, 0o644)
	if err != nil {
		log.Error().Str("filename", file).Msg("Writing back to file failed")

		return fmt.Errorf("writing to file %w", err)
	}

	return nil
}
