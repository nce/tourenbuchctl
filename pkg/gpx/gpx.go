package gpx

import (
	"fmt"
	"os"
	"time"

	"github.com/tkrajina/gpxgo/gpx"
)

type GPXPoint struct {
	Lat  float64
	Lon  float64
	Elev float64
	Time time.Duration
}

func CreateGPXFile(points []GPXPoint, startTime time.Time, name string, filename string, stravaId int64) error {

	g := gpx.GPX{
		Creator:     "tourenbuchctl",
		AuthorName:  "nce",
		Description: "Exported from Strava",
		AuthorLink:  "https://www.strava.com/activities/" + fmt.Sprintf("%d", stravaId),
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
		return err
	}

	return nil
}

func writeGPXToFile(g *gpx.GPX, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := g.ToXml(gpx.ToXmlParams{Indent: true})
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}
