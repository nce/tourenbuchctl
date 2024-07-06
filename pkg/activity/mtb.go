package activity

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/nce/tourenbuchctl/cmd/flags"
)

func PrintLines() {
	// Step 1: Generate lines and print them to stdout
	lines := []string{"line 1", "line 2", "line 3", "line 4"}

	// Step 2: Use fzf to let the user select a line
	cmd := exec.Command("fzf", "--tmux", "right,30%,40%", "--header", "Choose a activity")

	// Create pipes to communicate with fzf
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating stdin pipe:", err)
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating stdout pipe:", err)
		return
	}

	// Start the fzf process
	if err := cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "Error starting fzf:", err)
		return
	}

	// Write the lines to fzf's stdin
	go func() {
		defer stdin.Close()
		fmt.Println("Chose which activity")
		for _, line := range lines {
			fmt.Fprintln(stdin, line)
		}
	}()

	// Step 3: Capture the selected line from fzf's stdout
	var outBuf bytes.Buffer
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		outBuf.WriteString(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading from fzf:", err)
		return
	}

	// Wait for fzf to finish
	if err := cmd.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, "Error waiting for fzf:", err)
		return
	}

	// Print the selected line
	selectedLine := outBuf.String()
	fmt.Println("Selected line:", selectedLine)
}

func CreateActivity(flag *flags.CreateMtbFlags) error {

	mtb := &Activity{
		category:   "mtb",
		name:       flag.Core.Name,
		date:       flag.Core.Date,
		rating:     flag.Rating,
		difficulty: flag.Difficulty,
	}

	err := mtb.createFolder()
	if err != nil {
		panic("error creating folder")
	}

	for _, file := range []string{"description.md", "elevation.plt", "Makefile", "img-even.tex"} {

		text, err := mtb.initSkeleton(file)
		if err != nil {
			panic("error initializing skeleton")
		}

		file, err := os.Create(mtb.textLocation + "/" + file)
		if err != nil {
			log.Printf("Failed to create file: %v", err)
		}
		defer file.Close()

		_, err = file.WriteString(text)
		if err != nil {
			log.Printf("Failed to write to file: %v", err)
		}
	}

	return nil
}
