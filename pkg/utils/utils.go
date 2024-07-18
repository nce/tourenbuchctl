package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"golang.org/x/oauth2"
)

type Token struct {
	AccessToken  string    `json:"accessToken"`
	TokenType    string    `json:"tokenType"`
	RefreshToken string    `json:"refreshToken"`
	Expiry       time.Time `json:"expiry"`
}

func SaveToken(filename string, token *oauth2.Token) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("opening file: %s; error: %w", filename, err)
	}
	defer file.Close()

	data := &Token{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}

	return fmt.Errorf("encoding json: %w", json.NewEncoder(file).Encode(data))
}

func LoadToken(filename string) (*oauth2.Token, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening file %s; error: %w", filename, err)
	}
	defer file.Close()

	var data Token
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, fmt.Errorf("decoding token %s; error: %w", filename, err)
	}

	token := &oauth2.Token{
		AccessToken:  data.AccessToken,
		TokenType:    data.TokenType,
		RefreshToken: data.RefreshToken,
		Expiry:       data.Expiry,
	}

	return token, nil
}

func (t *Token) Valid() bool {
	return t.Expiry.After(time.Now())
}

func FuzzyFind(header string, input []string) (string, error) {
	cmd := exec.Command("fzf", "--tmux", "right,30%,40%", "--header", header)

	// Create pipes to communicate with fzf
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("creating stdin pipe %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("creating stdout pipe %w", err)
	}

	// Start the fzf process
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("starting fzf %w", err)
	}

	// Write the lines to fzf's stdin
	go func() {
		defer stdin.Close()

		for _, line := range input {
			fmt.Fprintln(stdin, line)
		}
	}()

	// Capture the selected line from fzf's stdout
	var outBuf bytes.Buffer

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		outBuf.WriteString(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading output from fzf %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("waiting for fzf output %w", err)
	}

	// Print the selected line
	selectedLine := outBuf.String()

	return selectedLine, nil
}
