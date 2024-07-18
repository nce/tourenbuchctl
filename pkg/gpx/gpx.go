package gpx

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/tkrajina/gpxgo/gpx"
)

type Point struct {
	Lat  float64
	Lon  float64
	Elev float64
	Time time.Duration
}

func CreateGPXFile(points []Point, startTime time.Time, name string, filename string, stravaID int64) error {
	g := gpx.GPX{
		Creator:     "tourenbuchctl",
		AuthorName:  "nce",
		Description: "Exported from Strava",
		AuthorLink:  "https://www.strava.com/activities/" + strconv.FormatInt(stravaID, 10),
		Name:        name,
		Time:        &startTime,
	}

	track := gpx.GPXTrack{}
	seg := gpx.GPXTrackSegment{}

	for _, point := range points {
		wpt := gpx.GPXPoint{
			Point: gpx.Point{
				Latitude:  point.Lat,
				Longitude: point.Lon,
				Elevation: *gpx.NewNullableFloat64(point.Elev),
			},
			Timestamp: startTime.Add(point.Time),
		}
		seg.AppendPoint(&wpt)
	}

	track.AppendSegment(&seg)
	g.AppendTrack(&track)

	err := writeGPXToFile(&g, filename)
	if err != nil {
		return fmt.Errorf("writing gpx to file %w", err)
	}

	return nil
}

func writeGPXToFile(g *gpx.GPX, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating file %s; error: %w", filename, err)
	}
	defer file.Close()

	data, err := g.ToXml(gpx.ToXmlParams{Indent: true})
	if err != nil {
		return fmt.Errorf("converting gpx to xml: %w", err)
	}

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("writing data to file: %w", err)
	}

	return nil
}
