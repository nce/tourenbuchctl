package strava

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nce/tourenbuchctl/pkg/gpx"
	api "github.com/nce/tourenbuchctl/pkg/stravaapi"
)

var ErrLatLongStreamMissing = errors.New("stream is not assigned")

func parseStreams(streams api.StreamSet) ([]gpx.Point, error) {
	if streams.Latlng == nil {
		return nil, fmt.Errorf("not set %w", ErrLatLongStreamMissing)
	}

	points := make([]gpx.Point, 0, len(streams.Latlng.Data))

	for i, latlng := range streams.Latlng.Data {
		point := gpx.Point{
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

func fetchActivityStream(activityID int64) (api.StreamSet, time.Time, string, error) {
	client, err := loginStrava()
	if err != nil {
		return api.StreamSet{}, time.Time{}, "", err
	}

	configuration := api.NewConfiguration()
	configuration.HTTPClient = client
	apiClient := api.NewAPIClient(configuration)

	keyTypes := []string{"latlng", "altitude", "time"}

	data, response, err := apiClient.StreamsApi.GetActivityStreams(context.Background(), activityID, keyTypes, true)
	if err != nil {
		return api.StreamSet{}, time.Time{}, "", fmt.Errorf("getActivityStreams(%d) failed: %w", activityID, err)
	}
	// Ensure the response body is closed
	defer func() {
		if response != nil && response.Body != nil {
			response.Body.Close()
		}
	}()

	meta, response, err := apiClient.ActivitiesApi.GetActivityById(
		context.Background(),
		activityID,
		&api.ActivitiesApiGetActivityByIdOpts{})
	if err != nil {
		return api.StreamSet{}, time.Time{}, "", fmt.Errorf("getActivityById(%d) failed: %w", activityID, err)
	}

	// Ensure the response body is closed
	defer func() {
		if response != nil && response.Body != nil {
			response.Body.Close()
		}
	}()

	return data, meta.StartDate, meta.Name, nil
}

func ExportStravaToGpx(activityID int64, filename string) error {
	data, startTime, name, err := fetchActivityStream(activityID)
	if err != nil {
		return err
	}

	gpxPoints, err := parseStreams(data)
	if err != nil {
		return fmt.Errorf("parse stream %w", err)
	}

	err = gpx.CreateGPXFile(gpxPoints, startTime, name, filename, activityID)
	if err != nil {
		return fmt.Errorf("create gpx file: %s; error %w", filename, err)
	}

	return nil
}
