package activity

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

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

func GetFromHeader[T any](dir string, field string) (T, error) { //nolint:ireturn
	var zero T

	data, err := os.ReadFile(dir + "/header.yaml")
	if err != nil {
		return zero, fmt.Errorf("error reading file: %w", err)
	}

	var act Header

	err = yaml.Unmarshal(data, &act)
	if err != nil {
		return zero, fmt.Errorf("error unmarshalling YAML: %w", err)
	}

	value, err := searchField[T](&act, field)
	if err != nil {
		return zero, fmt.Errorf("error searching field: %w", err)
	}

	return value, nil
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
