package utils

import (
	"fmt"
	"os"
	"testing"
)

func TestSplitDirectoryName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		dir  string
		name string
		date string
	}{
		{"3.laender-01.02.2001", "3.laender", "01.02.2001"},
	}

	for _, tt := range tests {
		t.Run(tt.dir+" dirs", func(t *testing.T) {
			t.Parallel()

			name, date, _ := SplitActivityDirectoryName(tt.dir)

			if name != tt.name {
				t.Errorf("got %s, want %s", name, tt.name)
			}

			if date != tt.date {
				t.Errorf("got %s, want %s", date, tt.date)
			}
		})
	}
}

func TestReadActivityTypeFromHeader(t *testing.T) {
	t.Parallel()

	dirName := t.TempDir()
	expected := "running"

	// Create a test header.yaml file in the testdata directory
	err := createTestHeaderFile(dirName, expected, "foo")
	if err != nil {
		t.Fatalf("error creating test file: %v", err)
	}

	// Test reading activity type from header.yaml
	activityType, err := ReadActivityTypeFromHeader(dirName)
	if err != nil {
		t.Fatalf("error reading activity type: %v", err)
	}

	if activityType != expected {
		t.Errorf("expected activity type %s, got %s", expected, activityType)
	}
}

func TestReadElevationProfileTypeFromHeader(t *testing.T) {
	t.Parallel()

	dirName := t.TempDir()
	expected := "left-axis"

	// Create a test header.yaml file in the testdata directory
	err := createTestHeaderFile(dirName, "foo", expected)
	if err != nil {
		t.Fatalf("error creating test file: %v", err)
	}

	// Test reading activity type from header.yaml
	elevationProfileType, err := ReadElevationProfileTypeFromHeader(dirName)
	if err != nil {
		t.Fatalf("error reading elevationProfileType: %v", err)
	}

	if elevationProfileType != expected {
		t.Errorf("expected layout elevationProfileType %s, got %s", expected, elevationProfileType)
	}
}

func createTestHeaderFile(dirName, activityType string, elevationProfileType string) error {
	headerContent := []byte(`
activity:
  type: ` + activityType + `
layout:
  elevationProfileType: ` + elevationProfileType + `
`)

	//nolint: gosec
	err := os.WriteFile(dirName+"/header.yaml", headerContent, 0o644)
	if err != nil {
		return fmt.Errorf("error setting up header.yaml: %w", err)
	}

	return nil
}
