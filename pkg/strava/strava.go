package strava

import (
	"context"
	"errors"
	"fmt"
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

var ErrStravaQuery = errors.New("querying strava for the activity failed")

const (
	//nolint:gosec
	tokenFile = "/tmp/stravatoken.json"
)

type Activity struct {
	ID          int64
	SportType   string
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

		log.Debug().Msgf("Oauth URL: %s", oauth.StravaOauthConfig.AuthCodeURL(oauth.OauthStateString))

		//nolint: gosec
		err := exec.Command("open", oauth.StravaOauthConfig.AuthCodeURL(oauth.OauthStateString)).Start()
		if err != nil {
			return nil, fmt.Errorf("exec open strava url: %w", err)
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

func FetchStravaData(date time.Time) (*Activity, error) {
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

	allActivites, response, err := apiClient.ActivitiesApi.GetLoggedInAthleteActivities(context.Background(), opts)
	if err != nil {
		return nil, fmt.Errorf("GetLoggedInAthleteActivities failed %w", err)
	}

	// Ensure the response body is closed
	defer func() {
		if response != nil && response.Body != nil {
			response.Body.Close()
		}
	}()

	// response from `GetActivityById`: DetailedActivity
	activityNames := []string{}
	for _, act := range allActivites {
		activityNames = append(activityNames, act.Name)
	}

	log.Debug().Str("Name", strings.Join(activityNames, " ")).Msg("Activities found on strava")

	name, err := utils.FuzzyFind("Select activities to sync", activityNames)
	if err != nil {
		return nil, fmt.Errorf("fuzzy find selected activities to sync: %w", err)
	}

	log.Debug().Str("Name", name).Msg("Activity selected via fuzzyFind")

	// select the activity called `name`
	for _, activitySummary := range allActivites {
		if activitySummary.Name == name {
			// querying for detailedActivity might be unnecessary, as the results
			// are already in the summary
			activity, response, err := apiClient.ActivitiesApi.GetActivityById(context.Background(), activitySummary.Id, nil)

			// Ensure the response body is closed
			defer func() {
				if response != nil && response.Body != nil {
					response.Body.Close()
				}
			}()

			if err != nil {
				return nil, fmt.Errorf("getactivitybyId(%d) failed: %w", activitySummary.Id, err)
			}

			return &Activity{
				Name:        activity.Name,
				SportType:   normalizeSportType(string(*activity.SportType)),
				Distance:    normalizeDistance(activity.Distance),
				StartDate:   activity.StartDate,
				MovingTime:  normalizeDuration(activity.MovingTime),
				ElapsedTime: normalizeDuration(activity.ElapsedTime),
				Ascent:      normalizeDistance(activity.TotalElevationGain),
				ID:          activitySummary.Id,
			}, nil
		}
	}

	return nil, fmt.Errorf("strava %w", ErrStravaQuery)
}

func normalizeDistance(distance float32) int {
	return int(distance)
}

func normalizeSportType(sportType string) string {
	switch sportType {
	case "MountainBikeRide":
		return "mtb"
	case "GravelRide":
		return "cyclo"
	}

	return sportType
}

func normalizeDuration(duration int32) time.Duration {
	return time.Duration(duration) * time.Second
}
