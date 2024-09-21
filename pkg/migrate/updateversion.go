package migrate

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

func normalizeHeaderFile(fileLocation string) error {
	if err := removeLinesFromFile(fileLocation + "header.yaml"); err != nil {
		return fmt.Errorf("error removing lines from header.yaml: %w", err)
	}

	return nil
}

func updateVersion(filename, newVersion string) error {
	originalContent, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	lines := strings.Split(string(originalContent), "\n")

	for i, line := range lines {
		if strings.Contains(line, "  version:") {
			lines[i] = "  version: " + newVersion
		}
	}

	updatedContent := strings.Join(lines, "\n")

	// Write the updated content back to the file
	//nolint: gosec
	err = os.WriteFile(filename, []byte(updatedContent), 0o644)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

func removeLinesFromFile(filePath string) error {
	linesToRemove := []string{
		"# vim: set filetype=yaml:",
		"---",
		"# default",
		"# adjust if excess of space",
	}

	inputFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer inputFile.Close()

	tempFile, err := os.CreateTemp("", "tempfile")
	if err != nil {
		return fmt.Errorf("error creating temp file: %w", err)
	}
	defer tempFile.Close()

	// Read the input file line by line and write to the temp file
	scanner := bufio.NewScanner(inputFile)
	writer := bufio.NewWriter(tempFile)

	for scanner.Scan() {
		line := scanner.Text()
		if !slices.Contains(linesToRemove, strings.TrimSpace(line)) {
			_, err := writer.WriteString(line + "\n")
			if err != nil {
				return fmt.Errorf("error writing to temp file: %w", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	writer.Flush()

	// Close both files
	inputFile.Close()
	tempFile.Close()

	// Replace the original file with the temp file
	err = os.Rename(tempFile.Name(), filePath)
	if err != nil {
		return fmt.Errorf("error replacing file: %w", err)
	}

	return nil
}

func InsertOrUpdateVersion(filePath string, version string) (bool, error) {
	versionLine := "meta:\n  version:"

	if err := normalizeHeaderFile(filePath); err != nil {
		return false, fmt.Errorf("error normalizingHeader: %w", err)
	}

	originalFile, err := os.ReadFile(filePath + "header.yaml")
	if err != nil {
		return false, fmt.Errorf("error reading file: %w", err)
	}

	if strings.HasPrefix(string(originalFile), versionLine) {
		err := updateVersion(filePath+"header.yaml", version)
		if err != nil {
			return false, fmt.Errorf("error setting new version on header.yaml: %w", err)
		}

		return true, nil
	}

	// Prepend the new content to the original content
	newContent := []byte(versionLine + " " + version + "\n" + string(originalFile))

	// Write the updated content back to the file
	//nolint: gosec
	err = os.WriteFile(filePath+"header.yaml", newContent, 0o644)
	if err != nil {
		return false, fmt.Errorf("error writing file: %w", err)
	}

	return true, nil
}
