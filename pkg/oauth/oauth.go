package oauth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nce/strava2tourenbuch/pkg/strava"
	"github.com/nce/strava2tourenbuch/pkg/utils"

	"golang.org/x/oauth2"
)

var (
	StravaOauthConfig = &oauth2.Config{
		ClientID:     "129561",
		ClientSecret: "87721aa9e6045c617f01d71dcb89026bfad3e961",
		Endpoint: oauth2.Endpoint{
			TokenURL: "https://www.strava.com/api/v3/oauth/token",
			AuthURL:  "https://www.strava.com/api/v3/oauth/authorize",
		},
		RedirectURL: "http://localhost:8080/callback",
		Scopes:      []string{"read,read_all,profile:read_all,activity:read_all"},
	}
	oauthStateString = "random" // This should be a random string for better security
)

func HandleStravaLogin(w http.ResponseWriter, r *http.Request) {
	url := StravaOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleStravaCallback(w http.ResponseWriter, r *http.Request, server *http.Server, tokenFile string) {
	if r.FormValue("state") != oauthStateString {
		log.Println("Invalid oauth state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	token, err := StravaOauthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		log.Println("Code exchange failed: ", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	client := StravaOauthConfig.Client(context.Background(), token)

	fmt.Fprint(w, "Login successful!")

	if err := utils.SaveToken(tokenFile, token); err != nil {
		log.Println("Failed to save token: ", err)
	}

	// Shutdown the server
	go func() {
		time.Sleep(1 * time.Second) // Give the response a moment to be sent
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
	}()

	strava.FetchStravaData(client)
}
