package activity

import (
	"bufio"
	"embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/nce/tourenbuchctl/pkg/utils"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// TODO:make this configurable via external config file (-> viper).
const (
	relativeTextLibraryPath  = "vcs/github/nce/tourenbuch/"
	relativeAssetLibraryPath = "Library/Mobile Documents/com~apple~CloudDocs/privat/sport/Tourenbuch/"
)

//go:embed templates/*
var content embed.FS

var ErrTourenbuchDirNameWrong = errors.New("directory name does not match expected schema")

type Meta struct {
	Category           string
	Name               string
	TextLocation       string
	AssetLocation      string
	StravaSync         bool
	StravaGpxSync      bool
	StravaID           int64
	QueryStartLocation bool
	Multiday           bool
}

type Tourenbuch struct {
	Title           string
	Date            time.Time
	Rating          int
	TrailDifficulty int
	SkiDifficulty   string
	AvalancheReport int
	StartLocationQr string
	Company         string
	Restaurant      string
	Distance        int
	Ascent          int
	MaxElevation    int
	MovingTime      time.Duration
	ElapsedTime     time.Duration
	StartTime       time.Time
}

type Activity struct {
	Meta Meta
	Tb   Tourenbuch
}

func (a *Activity) normalizeDate() string {
	return a.Tb.Date.Format("02.01.2006")
}

func (a *Activity) normalizeDateWithYear() string {
	return a.Tb.Date.Format("2006")
}

func normalizeDuration(d time.Duration) string {
	return fmt.Sprintf("%02d:%02d", int(d.Hours()), int(d.Minutes())%60)
}

func (a *Activity) normalizeStartTime() string {
	loc, _ := time.LoadLocation("Europe/Berlin")
	localTime := a.Tb.StartTime.In(loc)

	return localTime.Format("15:04")
}

func (a *Activity) normalizeDistance() string {
	return fmt.Sprintf("%.1f", float32(a.Tb.Distance)/float32(1000))
}

func normalizeElevation(elevation int) string {
	p := message.NewPrinter(language.German)

	return p.Sprintf("%d", elevation)
}

func (a *Activity) normalizeDateWithShortWeekday() string {
	date := a.Tb.Date.Format("02.01.2006")

	// abbrevations for german weekdays
	weekdayAbbreviations := map[string]string{
		"Sunday":    "So",
		"Monday":    "Mo",
		"Tuesday":   "Di",
		"Wednesday": "Mi",
		"Thursday":  "Do",
		"Friday":    "Fr",
		"Saturday":  "Sa",
	}

	fullWeekday := a.Tb.Date.Weekday().String()
	abbreviatedWeekday := weekdayAbbreviations[fullWeekday]

	return fmt.Sprintf("%s, %s", abbreviatedWeekday, date)
}

func GetTextLibraryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting Text Home Path: %w", err)
	}

	return fmt.Sprintf("%s/%s", home, relativeTextLibraryPath), nil
}

func GetAssetLibraryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting Library Home Path: %w", err)
	}

	return fmt.Sprintf("%s/%s", home, relativeAssetLibraryPath), nil
}

func GetActivityLocation() (string, string, error) {
	var loc string

	locations, err := getActivityLocations()
	if err != nil {
		return "", "", fmt.Errorf("get activity starting locations: %w", err)
	}

	loc, err = utils.FuzzyFind("Select Activity to update", locations)
	if err != nil {
		return "", "", fmt.Errorf("fuzzy finding activities to update: %w", err)
	}

	name, date, err := utils.SplitActivityDirectoryName(loc)
	if err != nil {
		return "", "", fmt.Errorf("directory name does not match tb pattern name-dd.mm.yyyy: %w", err)
	}

	return name, date, nil
}

func GetStartLocationQr() (string, error) {
	var loc string

	locations, err := getStartingLocations()
	if err != nil {
		return "", err
	}

	loc, err = utils.FuzzyFind("Select starting Location", locations)
	if err != nil {
		return "", fmt.Errorf("fuzzy finding starting location: %w", err)
	}

	if loc == "new" {
		newLoc, err := generateNewLocationQr()
		if err != nil {
			return "", fmt.Errorf("generating new location qr: %w", err)
		}

		return newLoc, nil
	}

	return loc, nil
}

func generateNewLocationQr() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	//nolint: forbidigo
	fmt.Print("Enter new location (47.123, 10.123): ")

	gpsLocation, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading gps location: %w", err)
	}

	gpsLocation = strings.TrimSpace(gpsLocation)

	//nolint: forbidigo
	fmt.Printf("Enter a name for the location: ")

	filename, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading filename: %w", err)
	}

	filename = strings.TrimSpace(filename)

	dir, err := GetTextLibraryPath()
	if err != nil {
		return "", err
	}

	dir += "/meta/location-qr"

	//nolint: gosec
	cmd := exec.Command("qrencode", "-t", "EPS", "-o", dir+"/"+filename+".eps", "geo:"+gpsLocation)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("running qrencode: %w", err)
	}

	log.Info().Str("filename", filename).Str("gpsLocation", gpsLocation).Msg("New QR-location created")

	return filename, nil
}

func getActivityLocations() ([]string, error) {
	var activityDirs []string

	dir, err := GetTextLibraryPath()
	if err != nil {
		return nil, fmt.Errorf("getting text library path: %s; error: %w", dir, err)
	}

	// Regular expression to match the schema "name-dd.mm.yyyy"
	regexPattern := regexp.MustCompile(`^[a-zA-Z0-9\.]+-\d{2}\.\d{2}\.\d{4}$`)

	err = filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if it's a directory and if the name matches the schema
		if info.IsDir() && regexPattern.MatchString(info.Name()) {
			activityDirs = append(activityDirs, info.Name())
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing filetree: %w", err)
	}

	return activityDirs, nil
}

func getStartingLocations() ([]string, error) {
	var epsFiles []string

	dir, err := GetTextLibraryPath()
	if err != nil {
		return nil, err
	}

	dir += "/meta/location-qr"

	// Read all existing starting locations
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading dir %s; error: %w", dir, err)
	}

	// Don't display file ending in the fzf list
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".eps" {
			fileName := strings.TrimSuffix(file.Name(), ".eps")
			epsFiles = append(epsFiles, fileName)
		}
	}

	return epsFiles, nil
}

func (a *Activity) getSeason() string {
	switch a.Tb.Date.Month() {
	case time.December, time.January, time.February, time.March:
		return "Winter"
	case time.April, time.May:
		return "Frühling"
	case time.June, time.July, time.August:
		return "Sommer"
	case time.September, time.October, time.November:
		return "Herbst"
	default:
		return "unknown"
	}
}

// checks if a string is the correct name of an activityType.
func ValidActivityType(activity string) bool {
	for _, validAcitvityType := range ActivityTypes {
		if activity == validAcitvityType.Name {
			return true
		}
	}

	return false
}
