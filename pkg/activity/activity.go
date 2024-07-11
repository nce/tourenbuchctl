package activity

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/rs/zerolog/log"

	"gopkg.in/yaml.v3"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/nce/tourenbuchctl/pkg/utils"
)

// TODO:  make this configurable via external config file (-> viper)
const (
	relativeTextLibraryPath  = "vcs/github/nce/tourenbuch/"
	relativeAssetLibraryPath = "Library/Mobile Documents/com~apple~CloudDocs/privat/sport/Tourenbuch/"
)

type Meta struct {
	Category           string
	Name               string
	TextLocation       string
	AssetLocation      string
	StravaSync         bool
	QueryStartLocation bool
}

type Tourenbuch struct {
	Title           string
	Date            time.Time
	Rating          int
	Difficulty      int
	StartLocationQr string
	Company         string
	Restaurant      string
	Distance        int
	Ascent          int
	MovingTime      time.Duration
	ElapsedTime     time.Duration
	StartTime       time.Time
}

type Activity struct {
	Meta Meta
	Tb   Tourenbuch
}

// Each Tourenbuch entry is represented by two folders. One folder contains the
// text part of the activity, the other one contains the assets (images, etc.).
// Text is stored in git; Assets are stored in iCloud
func (a *Activity) createFolder() error {

	textPath, err := getTextLibraryPath()
	if err != nil {
		return err
	}
	assetPath, err := getAssetLibraryPath()
	if err != nil {
		return err
	}

	dirs := [2]string{
		textPath + a.Meta.Category + "/" + a.Meta.Name + "-" + a.normalizeDate(),
		assetPath + a.Meta.Category + "/" + a.Meta.Name + "-" + a.normalizeDate() + "/" + "img",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	a.Meta.TextLocation = dirs[0]
	a.Meta.AssetLocation = dirs[1]

	return nil
}

func (a *Activity) initSkeleton(file string) (string, error) {
	location := fmt.Sprintf("templates/tourenbuch/%s/%s", a.Meta.Category, file)
	tmpl, err := template.ParseFiles(location)
	if err != nil {
		return "", err
	}

	textPath, err := getTextLibraryPath()
	if err != nil {
		return "", err
	}
	assetPath, err := getAssetLibraryPath()
	if err != nil {
		return "", err
	}

	data := struct {
		Name             string
		Date             string
		Stars            []struct{}
		Year             string
		AssetLibraryPath string
		TextLibraryPath  string
		StartLocationQr  string
		Title            string
		Company          string
		Difficulty       int
		Restaurant       string
		Season           string
	}{
		Name:             a.Meta.Name,
		Date:             a.normalizeDateWithShortWeekday(),
		Stars:            make([]struct{}, a.Tb.Rating),
		Year:             a.normalizeDateWithYear(),
		AssetLibraryPath: assetPath,
		TextLibraryPath:  textPath,
		StartLocationQr:  a.Tb.StartLocationQr,
		Title:            a.Tb.Title,
		Company:          a.Tb.Company,
		Difficulty:       a.Tb.Difficulty,
		Restaurant:       a.Tb.Restaurant,
		Season:           a.getSeason(),
	}

	io := new(strings.Builder)

	// Execute the template and write to the file
	err = tmpl.Execute(io, data)
	if err != nil {
		return "", err
	}
	return io.String(), nil
}

// Updates the existing stats-yaml structure with new data
func (a *Activity) updateActivity(file string) error {
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	// Parse the YAML file into a node tree
	var root yaml.Node
	err = yaml.Unmarshal(yamlFile, &root)
	if err != nil {
		return err
	}

	// No idea. This is written by AI
	// This modifies just one nested Yaml Node, without touching/killing the
	// rest of the file. It updates the statistic data (distance/ascent) of the activity

	// Navigate to the "stats" node and modify it
	// Traverse the document to find the "stats" key
	for i := 0; i < len(root.Content); i++ {
		if root.Content[i].Kind == yaml.MappingNode {
			for j := 0; j < len(root.Content[i].Content); j += 2 {
				keyNode := root.Content[i].Content[j]
				valueNode := root.Content[i].Content[j+1]

				if keyNode.Value == "stats" {
					// Modify the stats node
					// Example: Modify a specific key-value pair within the "stats" node
					for k := 0; k < len(valueNode.Content); k += 2 {
						keyNode := valueNode.Content[k]
						value := valueNode.Content[k+1]

						switch keyNode.Value {
						case "ascent":
							value.Value = a.normalizeAscent()
						case "distance":
							value.Value = a.normalizeDistance()
						case "movingTime":
							value.Value = normalizeDuration(a.Tb.MovingTime)
						case "overallTime":
							value.Value = normalizeDuration(a.Tb.ElapsedTime)
						case "startTime":
							value.Value = a.normalizeStartTime()
						}

						value.Tag = "!!str"

					}
				}
			}
		}
	}

	// Serialize the modified node tree back to a YAML string
	output, err := yaml.Marshal(&root)
	if err != nil {
		log.Fatal().Str("Error", err.Error()).Msg("Rading yaml file")
	}

	// Write the modified YAML string back to the file
	err = os.WriteFile(file, output, 0644)
	if err != nil {
		log.Fatal().Str("Error", err.Error()).Msg("Writing file")
	}

	return nil
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
	localTime := a.Tb.StartTime.Local()
	return localTime.Format("15:04")
}

func (a *Activity) normalizeDistance() string {
	return fmt.Sprintf("%.1f", float32(a.Tb.Distance)/float32(1000))
}

func (a *Activity) normalizeAscent() string {
	p := message.NewPrinter(language.German)
	return p.Sprintf("%d", a.Tb.Ascent)
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

func getTextLibraryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", home, relativeTextLibraryPath), nil
}

func getAssetLibraryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", home, relativeAssetLibraryPath), nil
}

func GetStartLocationQr() (string, error) {
	var loc string
	locations, err := getStartingLocations()
	if err != nil {
		return "", err
	}

	loc, err = utils.FuzzyFind("Select starting Location", locations)
	if err != nil {
		return "", err
	}

	return loc, nil
}

func getStartingLocations() ([]string, error) {
	var epsFiles []string

	dir, err := getTextLibraryPath()
	if err != nil {
		return nil, err
	}
	dir += "/meta/location-qr"

	// Read all existing starting locations
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
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
