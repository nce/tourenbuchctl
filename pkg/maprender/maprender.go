//nolint:varnamelen
package maprender

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg" // register JPEG tile decoder
	_ "image/png"  // register PNG tile decoder
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/fogleman/gg"
	"github.com/rs/zerolog/log"
	"github.com/tkrajina/gpxgo/gpx"
)

const (
	defaultWidth     = 1600
	defaultHeight    = 1000
	defaultTileSize  = 256
	defaultMinZoom   = 1
	defaultMaxZoom   = 15
	defaultPadding   = 40
	descentThreshold = 50
	defaultStyle     = "outdoors"
)

//nolint:gochecknoglobals
var (
	routeCasingColor  = color.RGBA{R: 44, G: 38, B: 31, A: 170}
	routeDefaultColor = color.RGBA{R: 234, G: 84, B: 42, A: 245}
	routeDescentColor = color.RGBA{R: 55, G: 145, B: 94, A: 255}
)

var (
	ErrMissingAPIKey        = errors.New("thunderforest api key is missing")
	ErrNoTrackPoints        = errors.New("gpx has no track points")
	ErrUnexpectedTileStatus = errors.New("unexpected thunderforest tile response status")
)

type Point struct {
	Lat          float64
	Lon          float64
	Elev         float64
	HasElevation bool
}

type Options struct {
	GPXPath    string
	OutputPath string
	APIKey     string
	Style      string
	Width      int
	Height     int
	HTTPClient *http.Client
}

type tileRange struct {
	MinX int
	MaxX int
	MinY int
	MaxY int
}

type pixelPoint struct {
	X float64
	Y float64
}

type viewport struct {
	Zoom int
	MinX float64
	MinY float64
}

func GenerateThunderforest(ctx context.Context, opts Options) error {
	if opts.APIKey == "" {
		return ErrMissingAPIKey
	}

	if opts.Width == 0 {
		opts.Width = defaultWidth
	}

	if opts.Height == 0 {
		opts.Height = defaultHeight
	}

	if opts.Style == "" {
		opts.Style = defaultStyle
	}

	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{Timeout: 20 * time.Second}
	}

	points, err := ReadGPXPoints(opts.GPXPath)
	if err != nil {
		return err
	}

	view := fitViewport(points, opts.Width, opts.Height, defaultPadding)
	tiles := tilesForViewport(view, opts.Width, opts.Height)

	canvas := image.NewRGBA(image.Rect(0, 0, opts.Width, opts.Height))
	background := &image.Uniform{C: color.RGBA{R: 239, G: 238, B: 232, A: 255}}
	draw.Draw(canvas, canvas.Bounds(), background, image.Point{}, draw.Src)

	for tileX := tiles.MinX; tileX <= tiles.MaxX; tileX++ {
		for tileY := tiles.MinY; tileY <= tiles.MaxY; tileY++ {
			tile, err := fetchThunderforestTile(ctx, opts.HTTPClient, opts.Style, view.Zoom, tileX, tileY, opts.APIKey)
			if err != nil {
				return fmt.Errorf("fetching thunderforest tile z%d/%d/%d: %w", view.Zoom, tileX, tileY, err)
			}

			drawTile(canvas, tile, tileX, tileY, view)
		}
	}

	dc := gg.NewContextForRGBA(canvas)
	markers := drawRoute(dc, points, view)
	drawAttribution(dc, opts.Width, opts.Height)
	log.Info().Int("directionMarkers", markers).Msg("Drew map route direction markers")

	if err := os.MkdirAll(filepath.Dir(opts.OutputPath), 0o755); err != nil {
		return fmt.Errorf("creating map output directory: %w", err)
	}

	if err := dc.SavePNG(opts.OutputPath); err != nil {
		return fmt.Errorf("saving map png: %w", err)
	}

	return nil
}

func GenerateForActivity(
	ctx context.Context,
	gpxPath string,
	outputPath string,
	apiKey string,
	activityType string,
) error {
	return GenerateThunderforest(ctx, Options{
		GPXPath:    gpxPath,
		OutputPath: outputPath,
		APIKey:     apiKey,
		Style:      StyleForActivity(activityType),
	})
}

func ReadGPXPoints(path string) ([]Point, error) {
	g, err := gpx.ParseFile(path)
	if err != nil {
		return nil, fmt.Errorf("parsing gpx file %s: %w", path, err)
	}

	points := make([]Point, 0)

	for _, track := range g.Tracks {
		for _, segment := range track.Segments {
			for _, point := range segment.Points {
				p := Point{
					Lat: point.Latitude,
					Lon: point.Longitude,
				}
				if point.Elevation.NotNull() {
					p.Elev = point.Elevation.Value()
					p.HasElevation = true
				}

				points = append(points, p)
			}
		}
	}

	if len(points) == 0 {
		return nil, ErrNoTrackPoints
	}

	return points, nil
}

func StyleForActivity(activityType string) string {
	switch activityType {
	case "mtb":
		return "cycle"
	case "skitour", "wandern":
		return defaultStyle
	default:
		return defaultStyle
	}
}

func thunderforestTileURL(style string, zoom int, x int, y int, apiKey string) string {
	u := url.URL{
		Scheme: "https",
		Host:   "api.thunderforest.com",
		Path:   fmt.Sprintf("/%s/%d/%d/%d.png", style, zoom, x, y),
	}

	q := u.Query()
	q.Set("apikey", apiKey)
	u.RawQuery = q.Encode()

	return u.String()
}

func fetchThunderforestTile(
	ctx context.Context,
	client *http.Client,
	style string,
	zoom int,
	x int,
	y int,
	apiKey string,
) (image.Image, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, thunderforestTileURL(style, zoom, x, y, apiKey), nil)
	if err != nil {
		return nil, fmt.Errorf("creating tile request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting tile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrUnexpectedTileStatus, resp.Status)
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decoding tile image: %w", err)
	}

	return img, nil
}

func fitViewport(points []Point, width int, height int, padding int) viewport {
	zoom := defaultMinZoom

	for z := defaultMaxZoom; z >= defaultMinZoom; z-- {
		minPoint, maxPoint := pixelBounds(points, z)
		if maxPoint.X-minPoint.X+float64(2*padding) <= float64(width) &&
			maxPoint.Y-minPoint.Y+float64(2*padding) <= float64(height) {
			zoom = z

			break
		}
	}

	minPoint, maxPoint := pixelBounds(points, zoom)
	centerX := (minPoint.X + maxPoint.X) / 2
	centerY := (minPoint.Y + maxPoint.Y) / 2

	return viewport{
		Zoom: zoom,
		MinX: centerX - float64(width)/2,
		MinY: centerY - float64(height)/2,
	}
}

func tilesForViewport(view viewport, width int, height int) tileRange {
	maxTile := (1 << view.Zoom) - 1

	return tileRange{
		MinX: clampTile(int(math.Floor(view.MinX/defaultTileSize)), maxTile),
		MaxX: clampTile(int(math.Floor((view.MinX+float64(width))/defaultTileSize)), maxTile),
		MinY: clampTile(int(math.Floor(view.MinY/defaultTileSize)), maxTile),
		MaxY: clampTile(int(math.Floor((view.MinY+float64(height))/defaultTileSize)), maxTile),
	}
}

func drawTile(canvas draw.Image, tile image.Image, x int, y int, view viewport) {
	destX := int(math.Round(float64(x*defaultTileSize) - view.MinX))
	destY := int(math.Round(float64(y*defaultTileSize) - view.MinY))
	dest := image.Rect(destX, destY, destX+defaultTileSize, destY+defaultTileSize)

	draw.Draw(canvas, dest, tile, tile.Bounds().Min, draw.Src)
}

func drawRoute(dc *gg.Context, points []Point, view viewport) int {
	if len(points) == 0 {
		return 0
	}

	dc.SetLineCap(gg.LineCapRound)
	dc.SetLineJoin(gg.LineJoinRound)

	descentSegments := descentSegmentMask(points)
	drawRouteStroke(dc, points, view, 8, routeCasingColor)
	drawStyledRouteStroke(dc, points, view, descentSegments, 4.5)
	markers := drawChevrons(dc, points, view, descentSegments)
	drawEndpoint(dc, projectToViewport(points[0], view), color.RGBA{R: 36, G: 133, B: 76, A: 255})
	drawEndpoint(dc, projectToViewport(points[len(points)-1], view), color.RGBA{R: 194, G: 54, B: 54, A: 255})

	return markers
}

func drawRouteStroke(dc *gg.Context, points []Point, view viewport, width float64, c color.Color) {
	first := projectToViewport(points[0], view)

	dc.NewSubPath()
	dc.MoveTo(first.X, first.Y)

	for _, point := range points[1:] {
		projected := projectToViewport(point, view)
		dc.LineTo(projected.X, projected.Y)
	}

	dc.SetColor(c)
	dc.SetLineWidth(width)
	dc.Stroke()
}

func drawStyledRouteStroke(dc *gg.Context, points []Point, view viewport, descentSegments []bool, width float64) {
	if len(points) < 2 {
		return
	}

	runStart := 0
	runColor := routeColorForSegment(descentSegments, 0)

	for segment := 1; segment < len(points)-1; segment++ {
		segmentColor := routeColorForSegment(descentSegments, segment)
		if segmentColor == runColor {
			continue
		}

		drawRouteRun(dc, points[runStart:segment+1], view, width, runColor)
		runStart = segment
		runColor = segmentColor
	}

	drawRouteRun(dc, points[runStart:], view, width, runColor)
}

func routeColorForSegment(descentSegments []bool, segment int) color.RGBA {
	if segment < len(descentSegments) && descentSegments[segment] {
		return routeDescentColor
	}

	return routeDefaultColor
}

func drawRouteRun(dc *gg.Context, points []Point, view viewport, width float64, c color.Color) {
	if len(points) < 2 {
		return
	}

	first := projectToViewport(points[0], view)

	dc.NewSubPath()
	dc.MoveTo(first.X, first.Y)

	for _, point := range points[1:] {
		projected := projectToViewport(point, view)
		dc.LineTo(projected.X, projected.Y)
	}

	dc.SetColor(c)
	dc.SetLineWidth(width)
	dc.Stroke()
}

func drawEndpoint(dc *gg.Context, point pixelPoint, c color.Color) {
	dc.DrawCircle(point.X, point.Y, 8)
	dc.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 230})
	dc.Fill()
	dc.DrawCircle(point.X, point.Y, 6)
	dc.SetColor(c)
	dc.Fill()
	dc.DrawCircle(point.X, point.Y, 8)
	dc.SetColor(color.RGBA{R: 32, G: 28, B: 24, A: 235})
	dc.SetLineWidth(2.5)
	dc.Stroke()
}

func descentSegmentMask(points []Point) []bool {
	if len(points) < 2 {
		return nil
	}

	mask := make([]bool, len(points)-1)
	runStart := 0
	runHigh := points[0].Elev
	runLow := points[0].Elev
	hasRun := points[0].HasElevation

	for i := 1; i < len(points); i++ {
		if !points[i].HasElevation {
			hasRun = false
			runStart = i

			continue
		}

		if !hasRun {
			runStart = i
			runHigh = points[i].Elev
			runLow = points[i].Elev
			hasRun = true

			continue
		}

		if points[i].Elev > runLow {
			if runHigh-runLow >= descentThreshold {
				for segment := runStart; segment < i-1; segment++ {
					mask[segment] = true
				}
			}

			runStart = i
			runHigh = points[i].Elev
			runLow = points[i].Elev

			continue
		}

		runLow = points[i].Elev
	}

	if hasRun && runHigh-runLow >= descentThreshold {
		for segment := runStart; segment < len(points)-1; segment++ {
			mask[segment] = true
		}
	}

	return mask
}

func drawChevrons(dc *gg.Context, points []Point, view viewport, descentSegments []bool) int {
	const (
		firstChevronAt = 120.0
		chevronSpacing = 260.0
		lookAhead      = 95.0
	)

	projected := make([]pixelPoint, 0, len(points))
	for _, point := range points {
		projected = append(projected, projectToViewport(point, view))
	}

	totalDistance := projectedDistance(projected)
	if totalDistance < 30 {
		return 0
	}

	targets := make([]float64, 0)
	for target := firstChevronAt; target < totalDistance; target += chevronSpacing {
		targets = append(targets, target)
	}

	if len(targets) == 0 {
		targets = append(targets, totalDistance/2)
	}

	for _, target := range targets {
		point := pointAtDistance(projected, target)
		angle := directionAtDistance(projected, target, lookAhead)
		c := colorAtDistance(projected, descentSegments, target)

		drawDirectionMarker(dc, point, angle, c)
	}

	return len(targets)
}

func directionAtDistance(points []pixelPoint, distance float64, lookAhead float64) float64 {
	from := pointAtDistance(points, math.Max(0, distance-lookAhead*0.35))
	to := pointAtDistance(points, distance+lookAhead)

	return math.Atan2(to.Y-from.Y, to.X-from.X)
}

func pointAtDistance(points []pixelPoint, target float64) pixelPoint {
	if len(points) == 0 {
		return pixelPoint{}
	}

	if target <= 0 {
		return points[0]
	}

	distance := 0.0

	for i := 1; i < len(points); i++ {
		start := points[i-1]
		end := points[i]
		segmentLength := math.Hypot(end.X-start.X, end.Y-start.Y)

		if segmentLength == 0 {
			continue
		}

		if distance+segmentLength >= target {
			ratio := (target - distance) / segmentLength

			return pixelPoint{
				X: start.X + (end.X-start.X)*ratio,
				Y: start.Y + (end.Y-start.Y)*ratio,
			}
		}

		distance += segmentLength
	}

	return points[len(points)-1]
}

func projectedDistance(points []pixelPoint) float64 {
	distance := 0.0

	for i := 1; i < len(points); i++ {
		distance += math.Hypot(points[i].X-points[i-1].X, points[i].Y-points[i-1].Y)
	}

	return distance
}

func colorAtDistance(points []pixelPoint, descentSegments []bool, target float64) color.Color {
	if len(points) < 2 {
		return routeDefaultColor
	}

	distance := 0.0

	for i := 1; i < len(points); i++ {
		start := points[i-1]
		end := points[i]
		segmentLength := math.Hypot(end.X-start.X, end.Y-start.Y)

		if segmentLength == 0 {
			continue
		}

		if distance+segmentLength >= target {
			if i-1 < len(descentSegments) && descentSegments[i-1] {
				return routeDescentColor
			}

			return routeDefaultColor
		}

		distance += segmentLength
	}

	if len(descentSegments) > 0 && descentSegments[len(descentSegments)-1] {
		return routeDescentColor
	}

	return routeDefaultColor
}

func drawDirectionMarker(dc *gg.Context, point pixelPoint, angle float64, c color.Color) {
	ux := math.Cos(angle)
	uy := math.Sin(angle)
	nx := -uy
	ny := ux

	drawSingleChevron(dc, point.X, point.Y, ux, uy, nx, ny, 24, 13, 5.5, color.RGBA{R: 44, G: 38, B: 31, A: 210})
	drawSingleChevron(dc, point.X, point.Y, ux, uy, nx, ny, 24, 13, 3.2, c)
}

func drawSingleChevron(
	dc *gg.Context,
	centerX float64,
	centerY float64,
	ux float64,
	uy float64,
	nx float64,
	ny float64,
	length float64,
	width float64,
	lineWidth float64,
	c color.Color,
) {
	tipX := centerX + ux*length*0.45
	tipY := centerY + uy*length*0.45
	baseX := centerX - ux*length*0.35
	baseY := centerY - uy*length*0.35
	leftX := baseX + nx*width*0.5
	leftY := baseY + ny*width*0.5
	rightX := baseX - nx*width*0.5
	rightY := baseY - ny*width*0.5

	dc.NewSubPath()
	dc.MoveTo(leftX, leftY)
	dc.LineTo(tipX, tipY)
	dc.LineTo(rightX, rightY)
	dc.SetColor(c)
	dc.SetLineWidth(lineWidth)
	dc.Stroke()
}

func drawAttribution(dc *gg.Context, width int, height int) {
	const text = "Maps (c) Thunderforest, Data (c) OpenStreetMap contributors"

	padding := 8.0
	textWidth, _ := dc.MeasureString(text)
	x := float64(width) - textWidth - padding
	y := float64(height) - padding

	dc.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 170})
	dc.DrawString(text, x+1, y+1)
	dc.SetColor(color.RGBA{R: 35, G: 35, B: 35, A: 180})
	dc.DrawString(text, x, y)
}

func pixelBounds(points []Point, zoom int) (pixelPoint, pixelPoint) {
	minPoint := latLonToPixel(points[0], zoom)
	maxPoint := minPoint

	for _, point := range points[1:] {
		pixel := latLonToPixel(point, zoom)
		minPoint.X = math.Min(minPoint.X, pixel.X)
		minPoint.Y = math.Min(minPoint.Y, pixel.Y)
		maxPoint.X = math.Max(maxPoint.X, pixel.X)
		maxPoint.Y = math.Max(maxPoint.Y, pixel.Y)
	}

	return minPoint, maxPoint
}

func projectToViewport(point Point, view viewport) pixelPoint {
	pixel := latLonToPixel(point, view.Zoom)

	return pixelPoint{
		X: pixel.X - view.MinX,
		Y: pixel.Y - view.MinY,
	}
}

func latLonToPixel(point Point, zoom int) pixelPoint {
	lat := math.Max(math.Min(point.Lat, 85.05112878), -85.05112878)
	sinLat := math.Sin(lat * math.Pi / 180)
	scale := float64(defaultTileSize) * math.Pow(2, float64(zoom))

	return pixelPoint{
		X: (point.Lon + 180) / 360 * scale,
		Y: (0.5 - math.Log((1+sinLat)/(1-sinLat))/(4*math.Pi)) * scale,
	}
}

func clampTile(value int, maxValue int) int {
	if value < 0 {
		return 0
	}

	if value > maxValue {
		return maxValue
	}

	return value
}
