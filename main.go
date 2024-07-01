package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/nce/strava2tourenbuch/pkg/oauth"
	"github.com/nce/strava2tourenbuch/pkg/strava"
	"github.com/nce/strava2tourenbuch/pkg/utils"
)

func main() {
	tokenFile := "/tmp/stravatoken.json"

	token, err := utils.LoadToken(tokenFile)
	if err == nil && token.Valid() {
		log.Println("Using existing token")
		client := oauth.StravaOauthConfig.Client(context.Background(), token)
		strava.FetchStravaData(client)
	} else {

		server := &http.Server{Addr: ":8080"}

		http.HandleFunc("/", handleMain)
		http.HandleFunc("/login", oauth.HandleStravaLogin)
		http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
			oauth.HandleStravaCallback(w, r, server, tokenFile)
		})

		log.Println("Started running on http://localhost:8080")
		log.Fatal(server.ListenAndServe())
	}
}

func handleMain(w http.ResponseWriter, r *http.Request) {
	var html = `<html><body><a href="/login">Log in with Strava</a></body></html>`
	fmt.Fprint(w, html)
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
