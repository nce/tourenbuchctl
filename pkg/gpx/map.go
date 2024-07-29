package gpx

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"os"

	"github.com/fogleman/gg"
	"github.com/tkrajina/gpxgo/gpx"
)

const mapboxStyle = "mapbox/outdoors-v12"
const imageWidth = 1024
const imageHeight = 800
const mapboxAccessToken = "pk.eyJ1IjoibmNlIiwiYSI6ImNsejZ0dnJreTAyaTcyanIzZ3VmZ201aHkifQ.li3srh1M7JPujaXHRBC04g"
const gpxFilePath = "/Users/nce//Library/Mobile Documents/com~apple~CloudDocs/privat/sport/Tourenbuch/mtb/multidaytrip/transalp-2024/day1/input.gpx"

const zoomLevel = 14

func fetchMapboxImage(lat, lon float64) (image.Image, error) {
	url := fmt.Sprintf("https://api.mapbox.com/styles/v1/%s/static/%f,%f,14/%dx%d?access_token=%s",
		mapboxStyle, lon, lat, imageWidth, imageHeight, mapboxAccessToken)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func CreateMap() {
	// Step 1: Read the GPX file
	gpxFile, err := os.Open(gpxFilePath)
	if err != nil {
		panic(err)
	}
	defer gpxFile.Close()

	gpxData, err := gpx.Parse(gpxFile)
	if err != nil {
		panic(err)
	}

	// Assume the GPX file contains at least one track with at least one segment
	if len(gpxData.Tracks) == 0 || len(gpxData.Tracks[0].Segments) == 0 || len(gpxData.Tracks[0].Segments[0].Points) == 0 {
		panic("No track data in GPX file")
	}

	points := gpxData.Tracks[0].Segments[0].Points
	centerLat := points[0].Latitude
	centerLon := points[0].Longitude

	// Step 2: Fetch the map image from Mapbox
	mapImage, err := fetchMapboxImage(centerLat, centerLon)
	if err != nil {
		panic(err)
	}

	// Step 3: Plot the GPX track on the image
	dc := gg.NewContextForImage(mapImage)
	dc.SetColor(color.RGBA{255, 0, 0, 255})
	dc.SetLineWidth(2)

	for i, point := range points {
		x := float64(imageWidth)*(point.Longitude-centerLon)/0.1 + float64(imageWidth)/2
		y := float64(imageHeight)*(centerLat-point.Latitude)/0.1 + float64(imageHeight)/2
		if i == 0 {
			dc.MoveTo(x, y)
		} else {
			dc.LineTo(x, y)
		}
	}
	dc.Stroke()

	// Step 4: Save the image to disk
	outputFile, err := os.Create("output.jpg")
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	err = jpeg.Encode(outputFile, dc.Image(), nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("Image saved successfully!")
}
