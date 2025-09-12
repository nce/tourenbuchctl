package pdfexport

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalExport struct {
	DestDirectory string
	DestFilename  string
}

func (l *LocalExport) Save(srcFile string) error {
	input, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("error opening source file: %w", err)
	}
	defer input.Close()

	destFilePath := filepath.Join(l.DestDirectory, l.DestFilename)

	output, err := os.Create(destFilePath)
	if err != nil {
		return fmt.Errorf("error creating export file: %w", err)
	}
	defer output.Close()

	_, err = io.Copy(output, input)
	if err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	err = output.Close()
	if err != nil {
		return fmt.Errorf("error closing file: %w", err)
	}

	return nil
}
