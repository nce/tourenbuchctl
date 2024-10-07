package oauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/nce/tourenbuchctl/pkg/templates"
	"github.com/nce/tourenbuchctl/pkg/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

const (
	authCallbackURL  = "http://localhost"
	authCallbackPort = 8080
	authCallbackPath = "/callback"
)

var (
	//nolint: gochecknoglobals
	StravaOauthConfig *oauth2.Config
	//nolint: gochecknoglobals
	OauthStateString = "random" // This should be a random string for better security

	ErrInvalidOauthState = errors.New("invalid oauth state")
)

func AuthStrava(tokenFile string) {
	server := runCallbackServer(handleStravaCallback(tokenFile))

	_, err := server()
	if err != nil {
		log.Error().Err(err).Msg("error creating client")
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
		Scopes:      []string{"read,read_all,profile:read_all,activity:read_all,activity:write"},
	}

	log.Debug().Str("clientsecret", StravaOauthConfig.ClientSecret).Send()
}

func handleStravaCallback(tokenFile string) func(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return func(w http.ResponseWriter, r *http.Request) (interface{}, error) {
		if r.FormValue("state") != OauthStateString {
			log.Error().Msg("Invalid oauth state")
			templates.HTMLRender(w, "templates/strava-failure.html")

			// insert html error here
			return nil, fmt.Errorf("%w: %s", ErrInvalidOauthState, r.FormValue("state"))
		}

		token, err := StravaOauthConfig.Exchange(context.Background(), r.FormValue("code"))
		if err != nil {
			log.Error().Err(err).Msg("Code exchange failed")
			templates.HTMLRender(w, "templates/strava-failure.html")

			return nil, fmt.Errorf("code exchange failed: %w", err)
		}

		templates.HTMLRender(w, "templates/strava-success.html")

		if err := utils.SaveToken(tokenFile, token); err != nil {
			log.Error().Err(err).Msg("Failed to save token")
		}

		return r.Body, nil
	}
}
