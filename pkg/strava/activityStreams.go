package strava

import (
	"context"
	"errors"
	"time"

	"github.com/nce/tourenbuchctl/pkg/gpx"

	api "github.com/nce/tourenbuchctl/pkg/stravaapi"
)

func parseStreams(streams api.StreamSet) ([]gpx.GPXPoint, error) {
	var points []gpx.GPXPoint

	if streams.Latlng == nil {
		return nil, errors.New("latlng stream is missing")
	}

	for i, latlng := range streams.Latlng.Data {
		point := gpx.GPXPoint{
			Lat: float64(latlng[0]),
			Lon: float64(latlng[1]),
		}
		if streams.Altitude != nil {
			point.Elev = float64(streams.Altitude.Data[i])
		}
		if streams.Time != nil {
			point.Time = time.Duration(streams.Time.Data[i])
		}
		points = append(points, point)
	}

	return points, nil
}

func fetchActivityStream(activityId int64) (api.StreamSet, time.Time, string, error) {

	client, err := loginStrava()
	if err != nil {
		return api.StreamSet{}, time.Time{}, "", err
	}
	configuration := api.NewConfiguration()
	configuration.HTTPClient = client
	apiClient := api.NewAPIClient(configuration)

	keyTypes := []string{"latlng", "altitude", "time"}

	data, _, err := apiClient.StreamsApi.GetActivityStreams(context.Background(), activityId, keyTypes, true)
	if err != nil {
		return api.StreamSet{}, time.Time{}, "", err
	}
	meta, _, err := apiClient.ActivitiesApi.GetActivityById(context.Background(), activityId, &api.ActivitiesApiGetActivityByIdOpts{})
	if err != nil {
		return api.StreamSet{}, time.Time{}, "", err
	}

	return data, meta.StartDate, meta.Name, nil
}

func ExportStravaToGpx(activityId int64, filename string) error {

	data, startTime, name, err := fetchActivityStream(activityId)
	if err != nil {
		return err
	}
	gpxPoints, err := parseStreams(data)
	if err != nil {
		return err
	}

	err = gpx.CreateGPXFile(gpxPoints, startTime, name, filename, activityId)
	if err != nil {
		return err
	}

	return nil

}
