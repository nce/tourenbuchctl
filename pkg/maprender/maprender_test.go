//nolint:varnamelen
package maprender

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fogleman/gg"
	"github.com/stretchr/testify/require"
)

func TestStyleForActivity(t *testing.T) {
	t.Parallel()

	require.Equal(t, "cycle", StyleForActivity("mtb"))
	require.Equal(t, "outdoors", StyleForActivity("skitour"))
	require.Equal(t, "outdoors", StyleForActivity("wandern"))
	require.Equal(t, "outdoors", StyleForActivity("unknown"))
}

func TestThunderforestTileURL(t *testing.T) {
	t.Parallel()

	url := thunderforestTileURL("outdoors", 12, 2170, 1420, "secret key")

	require.Equal(t, "https://api.thunderforest.com/outdoors/12/2170/1420.png?apikey=secret+key", url)
}

func TestLatLonToPixelAtZoomZero(t *testing.T) {
	t.Parallel()

	pixel := latLonToPixel(Point{Lat: 0, Lon: 0}, 0)

	require.InDelta(t, 128, pixel.X, 0.0001)
	require.InDelta(t, 128, pixel.Y, 0.0001)
}

func TestFitViewportKeepsRouteInsideCanvas(t *testing.T) {
	t.Parallel()

	points := []Point{
		{Lat: 47.35, Lon: 11.10},
		{Lat: 47.42, Lon: 11.21},
		{Lat: 47.39, Lon: 11.32},
	}

	view := fitViewport(points, 800, 500, 48)

	for _, point := range points {
		projected := projectToViewport(point, view)
		require.GreaterOrEqual(t, projected.X, 48.0)
		require.LessOrEqual(t, projected.X, 752.0)
		require.GreaterOrEqual(t, projected.Y, 48.0)
		require.LessOrEqual(t, projected.Y, 452.0)
	}
}

func TestReadGPXPoints(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "input.gpx")
	gpx := strings.TrimSpace(`<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1" creator="tourenbuchctl" xmlns="http://www.topografix.com/GPX/1/1">
  <trk>
    <trkseg>
      <trkpt lat="47.1" lon="11.1"><ele>1200</ele></trkpt>
      <trkpt lat="47.2" lon="11.2"><ele>1220</ele></trkpt>
    </trkseg>
  </trk>
</gpx>`)

	require.NoError(t, os.WriteFile(path, []byte(gpx), 0o600))

	points, err := ReadGPXPoints(path)

	require.NoError(t, err)
	require.Len(t, points, 2)
	require.Equal(t, Point{Lat: 47.1, Lon: 11.1, Elev: 1200, HasElevation: true}, points[0])
	require.Equal(t, Point{Lat: 47.2, Lon: 11.2, Elev: 1220, HasElevation: true}, points[1])
}

func TestDrawRouteIncludesVisibleChevrons(t *testing.T) {
	t.Parallel()

	points := []Point{
		{Lat: 47.35, Lon: 11.10},
		{Lat: 47.38, Lon: 11.17},
		{Lat: 47.42, Lon: 11.25},
		{Lat: 47.39, Lon: 11.34},
	}
	view := fitViewport(points, 800, 500, 40)
	dc := gg.NewContext(800, 500)
	dc.SetColor(color.White)
	dc.Clear()

	drawRoute(dc, points, view)

	require.Greater(t, countRouteOrangePixels(dc.Image()), 100)
}

func TestDrawChevronsHandlesDenseSubpixelSegments(t *testing.T) {
	t.Parallel()

	points := make([]Point, 0, 201)
	for i := range 201 {
		points = append(points, Point{
			Lat: 47.35,
			Lon: 11.10 + float64(i)*0.00001,
		})
	}

	view := fitViewport(points, 800, 500, 40)
	dc := gg.NewContext(800, 500)
	dc.SetColor(color.White)
	dc.Clear()

	markers := drawChevrons(dc, points, view, descentSegmentMask(points))

	require.Positive(t, markers)
}

func TestDescentSegmentMaskMarksRunsOverThreshold(t *testing.T) {
	t.Parallel()

	points := []Point{
		{Elev: 1100, HasElevation: true},
		{Elev: 1075, HasElevation: true},
		{Elev: 1040, HasElevation: true},
		{Elev: 1045, HasElevation: true},
		{Elev: 1030, HasElevation: true},
	}

	require.Equal(t, []bool{true, true, false, false}, descentSegmentMask(points))
}

func TestDescentSegmentMaskIgnoresSmallDescents(t *testing.T) {
	t.Parallel()

	points := []Point{
		{Elev: 1100, HasElevation: true},
		{Elev: 1075, HasElevation: true},
		{Elev: 1055, HasElevation: true},
	}

	require.Equal(t, []bool{false, false}, descentSegmentMask(points))
}

func countRouteOrangePixels(img image.Image) int {
	count := 0
	bounds := img.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if a > 0xf000 && r > 0xd000 && g > 0x3000 && g < 0x7000 && b < 0x4000 {
				count++
			}
		}
	}

	return count
}
