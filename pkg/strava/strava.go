package strava

import (
	"context"
	"fmt"
	"net/http"
	"os"

	api "github.com/nce/tourenbuchctl/pkg/stravaapi"
)

func FetchStravaData(client *http.Client) {
	configuration := api.NewConfiguration()
	configuration.HTTPClient = client
	apiClient := api.NewAPIClient(configuration)

	id := int64(11769165697) // int64 | The identifier of the activity.
	foo, re, err := apiClient.ActivitiesApi.GetActivityById(context.Background(), id, nil)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ActivitiesAPI.xxx`: %v\n\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", re)
	}
	// response from `GetActivityById`: DetailedActivity
	fmt.Fprintf(os.Stdout, "%s war %f lang\n", foo.Name, foo.Distance)
}
