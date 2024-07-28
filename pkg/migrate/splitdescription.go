package migrate

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// migrate "description.md" to header.yaml and description.md.
func SplitDescriptionFile(textDir string) (bool, error) {
	// Check if header.yaml already exists
	if _, err := os.Stat(textDir + "header.yaml"); err == nil {
		return false, nil
	}

	file, err := os.Open(textDir + "description.md")
	if err != nil {
		return false, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Read the file contents
	scanner := bufio.NewScanner(file)
	splitFound := false

	var headerLines, descriptionLines []string

	for scanner.Scan() {
		line := scanner.Text()
		if line == "..." {
			splitFound = true

			continue
		}

		if splitFound {
			descriptionLines = append(descriptionLines, line)
		} else {
			headerLines = append(headerLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error reading file: %w", err)
	}

	// Write the header part to header.yaml
	//nolint: gosec
	err = os.WriteFile(textDir+"header.yaml", []byte(strings.Join(headerLines, "\n")), 0o644)
	if err != nil {
		return false, fmt.Errorf("error writing header.yaml: %w", err)
	}

	// Write the description part to description.md
	//nolint: gosec
	err = os.WriteFile(textDir+"description.md", []byte(strings.Join(descriptionLines, "\n")), 0o644)
	if err != nil {
		return false, fmt.Errorf("error writing description.md: %w", err)
	}

	return true, nil
}
