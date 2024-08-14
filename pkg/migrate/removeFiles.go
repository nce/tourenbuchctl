package migrate

import (
	"fmt"
	"os"
)

// Remove old tex artefacts; Makefiles etc.
func RemoveObsoleteFiles(textDir string) (bool, error) {
	files := []string{
		"Makefile",
		"elevation.tex",
		"elevation.pdf",
		"gpxdata.txt",
		"description.aux",
		"description.tex",
	}

	for _, filename := range files {
		file := textDir + filename
		if _, err := os.Stat(file); err == nil {
			err := os.Remove(file)
			if err != nil {
				return false, fmt.Errorf("error deleting file: %w", err)
			}
		}
	}

	return true, nil
}
