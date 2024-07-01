package utils

import (
	"encoding/json"
	"os"
	"time"

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
