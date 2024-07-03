package main

import (
	"github.com/nce/tourenbuchctl/cmd"
)

func main() {
	cmd.Execute()
}

// 	id := int64(11769165697) // int64 | The identifier of the activity.
//
// 	configuration := strava.NewConfiguration()
//
// 	tokenSource := oauth2cfg.TokenSource(createContext(httpClient), &token)
// 	auth := context.WithValue(oauth2.NoContext, sw.ContextOAuth2, tokenSource)
// 	r, err := client.Service.Operation(auth, args)
// 	//	ctx := context.WithValue(context.Background(), "ContextAPIKey", "6ce1da3")
// 	//	apiClient := strava.NewAPIClient(configuration)
//
// 	resp, r, err := apiClient.ActivitiesApi.GetActivityById(ctx, id, nil)
//
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Error when calling `ActivitiesAPI.GetActivityById``: %v\n", err)
// 		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
// 	}
// 	// response from `GetActivityById`: DetailedActivity
// 	fmt.Fprintf(os.Stdout, "Response from `ActivitiesAPI.GetActivityById`: %v\n", resp)
// }
