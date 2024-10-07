package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"time"

	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

type Token struct {
	AccessToken  string    `json:"accessToken"`
	TokenType    string    `json:"tokenType"`
	RefreshToken string    `json:"refreshToken"`
	Expiry       time.Time `json:"expiry"`
}

type Header struct {
	Activity Activity `yaml:"activity"`
	Layout   Layout   `yaml:"layout"`
}

type Activity struct {
	Type string `yaml:"type"`
}

type Layout struct {
	ElevationProfileType string `yaml:"elevationProfileType"`
}

var (
	ErrActivityTypeEmpty               = errors.New("Activity.Type is empty")
	ErrLayoutElevationProfileTypeEmpty = errors.New("Layout.ElevationProfileType is empty")
)

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

	err = json.NewEncoder(file).Encode(data)
	if err != nil {
		return fmt.Errorf("encoding json: %w", err)
	}

	return nil
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

// expects only the last part of the path.
func SplitActivityDirectoryName(dirName string) (string, string, error) {
	// Regular expression to match the schema "name-dd.mm.yyyy"
	regexPattern := regexp.MustCompile(`^([a-zA-Z0-9\.]+)-(\d{2}\.\d{2}\.\d{4})$`)

	matches := regexPattern.FindStringSubmatch(dirName)
	if matches == nil {
		return "", "", fmt.Errorf("directory name %q does not match %w", dirName, ErrTourenbuchDirNameWrong)
	}

	// The first submatch is the full match, the second is the name part, and the third is the date string
	namePart := matches[1]
	datePart := matches[2]

	return namePart, datePart, nil
}

func ReadActivityTypeFromHeader(dirName string) (string, error) {
	data, err := os.ReadFile(dirName + "/header.yaml")
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	var act Header

	err = yaml.Unmarshal(data, &act)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling YAML: %w", err)
	}

	if act.Activity.Type == "" {
		return "", fmt.Errorf("parsing %s/header.yaml: %w", dirName, ErrActivityTypeEmpty)
	}

	return act.Activity.Type, nil
}

func ReadElevationProfileTypeFromHeader(dirName string) (string, error) {
	data, err := os.ReadFile(dirName + "/header.yaml")
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	var act Header

	err = yaml.Unmarshal(data, &act)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling YAML: %w", err)
	}

	if act.Layout.ElevationProfileType == "" {
		return "", fmt.Errorf("parsing %s/header.yaml: %w", dirName, ErrLayoutElevationProfileTypeEmpty)
	}

	return act.Layout.ElevationProfileType, nil
}
