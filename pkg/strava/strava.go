package strava

import (
	"context"
	"fmt"
	"net/http"
	"os"

	api "github.com/nce/strava2tourenbuch/pkg/stravaapi"
)

func FetchStravaData(client *http.Client) {
	configuration := api.NewConfiguration()
	//bearer := "Bearer " + token.AccessToken
	//configuration.AddDefaultHeader("Authorization", bearer)
	configuration.HTTPClient = client
	//fmt.Fprintf(os.Stdout, "%v\n", configuration)
	apiClient := api.NewAPIClient(configuration)

	id := int64(11769165697) // int64 | The identifier of the activity.
	//ctx := context.WithValue(oauth2.NoContext, strava.ContextOAuth2, token)
	foo, re, err := apiClient.ActivitiesApi.GetActivityById(context.Background(), id, nil)
	//foo, re, err := apiClient.ActivitiesApi.GetLoggedInAthleteActivities(context.Background(), nil)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ActivitiesAPI.xxx`: %v\n\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", re)
	}
	// response from `GetActivityById`: DetailedActivity
	fmt.Fprintf(os.Stdout, "%s war %f lang\n", foo.Name, foo.Distance)
}
