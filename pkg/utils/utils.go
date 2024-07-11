package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

type Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
}

func SaveToken(filename string, token *oauth2.Token) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	data := &Token{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}

	return json.NewEncoder(f).Encode(data)
}

func LoadToken(filename string) (*oauth2.Token, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var data Token
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		return nil, err
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
		log.Error().Err(err).Msg("Error creating stdin pipe")
		return "", err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Error().Err(err).Msg("Error creating stdout pipe")
		return "", err
	}

	// Start the fzf process
	if err := cmd.Start(); err != nil {
		log.Error().Err(err).Msg("Error starting fzf")
		return "", err
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
		log.Error().Err(err).Msg("Error reading from fzf output")
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		log.Error().Err(err).Msg("Error waiting for fzf output")
		return "", err
	}

	// Print the selected line
	selectedLine := outBuf.String()
	return selectedLine, nil
}
