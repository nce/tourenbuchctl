package strava

import (
	"context"
	"errors"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/antihax/optional"
	"github.com/nce/tourenbuchctl/pkg/oauth"
	api "github.com/nce/tourenbuchctl/pkg/stravaapi"
	"github.com/nce/tourenbuchctl/pkg/utils"
	"github.com/rs/zerolog/log"
)

const (
	tokenFile = "/tmp/stravatoken.json"
)

type StravaActivity struct {
	Name        string
	Distance    int
	Ascent      int
	StartDate   time.Time
	MovingTime  time.Duration
	ElapsedTime time.Duration
}

func loginStrava() (*http.Client, error) {

	client := &http.Client{}

	token, err := utils.LoadToken(tokenFile)
	if err == nil && token.Valid() {
		log.Debug().Msg("Using existing token from tokenfile to query strava")
		client = oauth.StravaOauthConfig.Client(context.Background(), token)
	} else {

		log.Debug().Msg("Refreshing token from strava")
		oauth.InitStravaOauthConfig()

		log.Debug().Str("Oauth URL", oauth.StravaOauthConfig.AuthCodeURL(oauth.OauthStateString))

		err := exec.Command("open", oauth.StravaOauthConfig.AuthCodeURL(oauth.OauthStateString)).Start()
		if err != nil {
			return nil, err
		}

		oauth.AuthStrava(tokenFile)

		token, err := utils.LoadToken(tokenFile)

		if err == nil && token.Valid() {
			log.Debug().Msg("Using newly acquired token to query strava")
			client = oauth.StravaOauthConfig.Client(context.Background(), token)
		}
	}

	return client, nil

}

func FetchStravaData(date time.Time) (*StravaActivity, error) {
	client, err := loginStrava()
	if err != nil {
		return nil, err
	}
	configuration := api.NewConfiguration()
	configuration.HTTPClient = client
	apiClient := api.NewAPIClient(configuration)

	opts := &api.ActivitiesApiGetLoggedInAthleteActivitiesOpts{
		Before: optional.NewInt32(int32(date.Add(24 * time.Hour).Unix())),
		After:  optional.NewInt32(int32(date.Unix())),
	}
	allActivites, _, err := apiClient.ActivitiesApi.GetLoggedInAthleteActivities(context.Background(), opts)

	if err != nil {
		return nil, err
	}

	// response from `GetActivityById`: DetailedActivity
	activityNames := []string{}
	for _, act := range allActivites {
		activityNames = append(activityNames, act.Name)
	}
	log.Debug().Str("Name", strings.Join(activityNames, " ")).Msg("Activities found on strava")

	name, err := utils.FuzzyFind("Select activities to sync", activityNames)
	if err != nil {
		return nil, err
	}
	log.Debug().Str("Name", name).Msg("Activity selected via fuzzyFind")

	// select the activity called `name`
	for _, activitySummary := range allActivites {
		if activitySummary.Name == name {

			// querying for detailedActivity might be unnecessary, as the results
			// are already in the summary
			activity, _, err := apiClient.ActivitiesApi.GetActivityById(context.Background(), activitySummary.Id, nil)
			if err != nil {
				return nil, err
			}

			return &StravaActivity{
				Name:        activity.Name,
				Distance:    normalizeDistance(activity.Distance),
				StartDate:   activity.StartDate,
				MovingTime:  normalizeDuration(activity.MovingTime),
				ElapsedTime: normalizeDuration(activity.ElapsedTime),
				Ascent:      normalizeDistance(activity.TotalElevationGain),
			}, nil
		}
	}

	return nil, errors.New("Querying strava for the activity failed")
}

func normalizeDistance(distance float32) int {
	return int(distance)
}

func normalizeDuration(duration int32) time.Duration {
	return time.Duration(duration) * time.Second
}
