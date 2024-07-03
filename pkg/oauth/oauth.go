package oauth

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/nce/strava2tourenbuch/pkg/utils"
	"github.com/spf13/viper"

	"golang.org/x/oauth2"
)

const (
	authCallbackURL  = "http://localhost"
	authCallbackPort = 8080
	authCallbackPath = "/callback"
)

var (
	StravaOauthConfig *oauth2.Config
	OauthStateString  = "random" // This should be a random string for better security
)

func AuthStrava(tokenFile string) {

	server := runCallbackServer(handleStravaCallback(tokenFile))
	_, err := server()
	if err != nil {
		log.Fatalf("error creating client")
	}

}

func InitStravaOauthConfig() {
	StravaOauthConfig = &oauth2.Config{
		ClientID:     viper.GetString("STRAVA_CLIENT_ID"),
		ClientSecret: viper.GetString("STRAVA_CLIENT_SECRET"),
		Endpoint: oauth2.Endpoint{
			TokenURL: "https://www.strava.com/api/v3/oauth/token",
			AuthURL:  "https://www.strava.com/api/v3/oauth/authorize",
		},
		RedirectURL: fmt.Sprintf("%s:%d%s", authCallbackURL, authCallbackPort, authCallbackPath),
		Scopes:      []string{"read,read_all,profile:read_all,activity:read_all"},
	}

	log.Println(StravaOauthConfig.ClientSecret)
}

func handleStravaCallback(tokenFile string) func(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return func(w http.ResponseWriter, r *http.Request) (interface{}, error) {
		if r.FormValue("state") != OauthStateString {
			log.Println("Invalid oauth state")
			// insert html error here
			return nil, fmt.Errorf("foo")
		}

		token, err := StravaOauthConfig.Exchange(context.Background(), r.FormValue("code"))
		if err != nil {
			log.Println("Code exchange failed: ", err)
			// insert html error here
			return nil, fmt.Errorf("bar")
		}

		//	client := StravaOauthConfig.Client(context.Background(), token)

		fmt.Fprint(w, "Login successful!")

		if err := utils.SaveToken(tokenFile, token); err != nil {
			log.Println("Failed to save token: ", err)
		}

		return r.Body, nil
	}

}
