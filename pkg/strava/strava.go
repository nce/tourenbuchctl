package strava

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/antihax/optional"
	"github.com/nce/tourenbuchctl/pkg/oauth"
	api "github.com/nce/tourenbuchctl/pkg/stravaapi"
	"github.com/nce/tourenbuchctl/pkg/utils"
)

const (
	tokenFile = "/tmp/stravatoken.json"
)

type StravaActivity struct {
	Name        string
	Distance    string
	Ascent      string
	StartDate   string
	MovingTime  string
	ElapsedTime string
}

func loginStrava() *http.Client {

	client := &http.Client{}

	token, err := utils.LoadToken(tokenFile)
	if err == nil && token.Valid() {
		log.Println("Using existing token")
		client = oauth.StravaOauthConfig.Client(context.Background(), token)
	} else {

		oauth.InitStravaOauthConfig()

		log.Println("Using no token")
		log.Println("Sent to:", oauth.StravaOauthConfig.AuthCodeURL(oauth.OauthStateString))
		err := exec.Command("open", oauth.StravaOauthConfig.AuthCodeURL(oauth.OauthStateString)).Start()
		if err != nil {
			log.Fatal(err)
		}

		oauth.AuthStrava(tokenFile)

		token, err := utils.LoadToken(tokenFile)

		if err == nil && token.Valid() {
			log.Println("Using new token")
			client = oauth.StravaOauthConfig.Client(context.Background(), token)
		}
	}

	return client

}

func FetchStravaData(date time.Time) *StravaActivity {
	client := loginStrava()
	configuration := api.NewConfiguration()
	configuration.HTTPClient = client
	apiClient := api.NewAPIClient(configuration)

	opts := &api.ActivitiesApiGetLoggedInAthleteActivitiesOpts{
		//Before: optional.NewInt32(int32(date.AddDate(0, 0, 1).Unix())),
		Before: optional.NewInt32(int32(date.Add(24 * time.Hour).Unix())),
		After:  optional.NewInt32(int32(date.Unix())),
	}
	foo, re, err := apiClient.ActivitiesApi.GetLoggedInAthleteActivities(context.Background(), opts)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ActivitiesAPI.xxx`: %v\n\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", re)
	}
	// response from `GetActivityById`: DetailedActivity
	activityNames := []string{}
	for _, act := range foo {
		activityNames = append(activityNames, act.Name)
	}

	name, err := utils.FuzzyFind("Select activities to sync", activityNames)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't find activity to sync %s", err)
	}

	// select the activity called `name`
	for _, activitySummary := range foo {
		if activitySummary.Name == name {

			activity, _, err := apiClient.ActivitiesApi.GetActivityById(context.Background(), activitySummary.Id, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error when calling `ActivitiesAPI.GetActivityById`: %v\n\n", err)
			}

			return &StravaActivity{
				Name:        activity.Name,
				Distance:    normalizeDistance(activity.Distance),
				StartDate:   normalizeStartDate(activity.StartDate),
				MovingTime:  normalizeElapsedTime(activity.MovingTime),
				ElapsedTime: normalizeElapsedTime(activity.ElapsedTime),
				Ascent:      fmt.Sprintf("%.0f", activity.TotalElevationGain),
			}
		}
	}

	return nil
}

func normalizeDistance(distance float32) string {
	return fmt.Sprintf("%d", int32(distance/1000))
}

func normalizeStartDate(startDate time.Time) string {
	localTime := startDate.Local()
	return localTime.Format("15:04")
}

func normalizeElapsedTime(seconds int32) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	return fmt.Sprintf("%02d:%02d", hours, minutes)

}
