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

	var file *os.File

	file, err := os.Open(textDir + "description.md")
	if err != nil {
		// check if theres a german version of the file
		file, err = os.Open(textDir + "beschreibung.md")
		if err != nil {
			return false, fmt.Errorf("error opening file: %w", err)
		}
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

func SplitImagesIncludeInOwnFile(textDir string) (bool, error) {
	// check if file already exists, assume it's setup correct
	if _, err := os.Stat(textDir + "images.tex"); err == nil {
		return false, nil
	}

	sourceFile, err := os.Open(textDir + "description.md")
	if err != nil {
		return false, fmt.Errorf("error opening file: %w", err)
	}
	defer sourceFile.Close()

	// create a temporary file for the modified description file
	tempFile, err := os.CreateTemp("", "description_temp_*.md")
	if err != nil {
		return false, fmt.Errorf("error creating temp file: %w", err)
	}
	defer tempFile.Close()

	scanner := bufio.NewScanner(sourceFile)

	var insideFigureBlock bool

	var figureContent []string

	var destFile *os.File

	for scanner.Scan() {
		line := scanner.Text()

		// Check if line starts with \begin{figure}[b!]
		if strings.HasPrefix(line, `\begin{figure}[b!]`) {
			insideFigureBlock = true

			figureContent = append(figureContent, `\hfill`+"\n")
			figureContent = append(figureContent, line)

			destFile, err = os.Create("images.tex")
			if err != nil {
				return false, fmt.Errorf("error creating destination file: %w", err)
			}
			defer destFile.Close()

			continue
		}

		// If inside the figure block, collect the lines but don't write them to the temp file
		//nolint: nestif
		if insideFigureBlock {
			figureContent = append(figureContent, line)

			if strings.HasPrefix(line, `\end{figure}`) {
				insideFigureBlock = false
				// Write the figure block to images.tex
				for _, figLine := range figureContent {
					_, err := destFile.WriteString(figLine + "\n")
					if err != nil {
						return false, fmt.Errorf("error writing image content to images.tex: %w", err)
					}
				}

				_, err := destFile.WriteString("\n") // Add a blank line after each figure block
				if err != nil {
					return false, fmt.Errorf("error writing newline to new file: %w", err)
				}

				figureContent = nil
			}

			continue
		}

		// Write non-figure block lines to the temporary file
		// remove all \hfill attributes
		if line != `\hfill` {
			_, err := tempFile.WriteString(line + "\n")
			if err != nil {
				return false, fmt.Errorf("error writing existing content to tempfile: %w", err)
			}
		}
	}

	if insideFigureBlock {
		_, err := tempFile.WriteString("\n" + `\input{\textpath/images}`)
		if err != nil {
			return false, fmt.Errorf("error writing image include to description: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error reading file: %w", err)
	}

	tempFile.Close()

	// replace the original file with the updated temp file
	if err := os.Rename(tempFile.Name(), textDir+"description.md"); err != nil {
		return false, fmt.Errorf("error replacing the original file: %w", err)
	}

	return true, nil
}
