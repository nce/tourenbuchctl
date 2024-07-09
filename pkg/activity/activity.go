package activity

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/nce/tourenbuchctl/cmd/flags"
	"github.com/nce/tourenbuchctl/pkg/utils"
)

// TODO:  make this configurable via external config file (-> viper)
const (
	relativeTextLibraryPath  = "vcs/github/nce/tourenbuch/"
	relativeAssetLibraryPath = "Library/Mobile Documents/com~apple~CloudDocs/privat/sport/Tourenbuch/"
)

type Activity struct {
	category        string
	name            string
	title           string
	textLocation    string
	assetLocation   string
	date            time.Time
	rating          int
	difficulty      int
	startLocationQr string
	company         string
	restaurant      string
	distance        string
	ascent          string
	movingTime      string
	elapsedTime     string
	startTime       string
}

type ActivityClasses interface {
	CreateActivity(flag *flags.CreateFlags) error
}

// Each Tourenbuch entry is represented by two folders. One folder contains the
// text part of the activity, the other one contains the assets (images, etc.).
// Text is stored in git; Assets are stored in iCloud
func (a *Activity) createFolder() error {

	dirs := [2]string{
		getTextLibraryPath() + a.category + "/" + a.name + "-" + a.normalizeDate(),
		getAssetLibraryPath() + a.category + "/" + a.name + "-" + a.normalizeDate() + "/" + "img",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Println("error creating folder: ", err)
		}
	}

	a.textLocation = dirs[0]
	a.assetLocation = dirs[1]

	return nil
}

func (a *Activity) initSkeleton(file string) (string, error) {
	location := fmt.Sprintf("templates/tourenbuch/%s/%s", a.category, file)
	tmpl, err := template.ParseFiles(location)
	if err != nil {
		log.Fatalf("Failed to parse template file: %v", err)
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
		Name:             a.name,
		Date:             a.normalizeDateWithShortWeekday(),
		Stars:            make([]struct{}, a.rating),
		Year:             a.normalizeDateWithYear(),
		AssetLibraryPath: getAssetLibraryPath(),
		TextLibraryPath:  getTextLibraryPath(),
		StartLocationQr:  a.startLocationQr,
		Title:            a.title,
		Company:          a.company,
		Difficulty:       a.difficulty,
		Restaurant:       a.restaurant,
		Season:           a.getSeason(),
	}

	io := new(strings.Builder)

	// Execute the template and write to the file
	err = tmpl.Execute(io, data)
	if err != nil {
		panic(err)
	}
	return io.String(), nil
}

func (a *Activity) updateActivity(file string) {
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}

	// Parse the YAML file into a node tree
	var root yaml.Node
	err = yaml.Unmarshal(yamlFile, &root)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// No idea. This is written by AI
	// This modifies just one nested Yaml Node, without touching/killing the
	// rest of the file. It updates the statistic data

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
							value.Value = a.ascent
						case "distance":
							value.Value = a.distance
						case "movingTime":
							value.Value = a.movingTime
						case "overallTime":
							value.Value = a.elapsedTime
						case "startTime":
							value.Value = a.startTime
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
		log.Fatalf("error: %v", err)
	}

	// Write the modified YAML string back to the file
	err = os.WriteFile(file, output, 0644)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

}

func (a *Activity) normalizeDate() string {
	return a.date.Format("02.01.2006")
}

func (a *Activity) normalizeDateWithYear() string {
	return a.date.Format("2006")
}

func (a *Activity) normalizeDateWithShortWeekday() string {

	date := a.date.Format("02.01.2006")

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

	fullWeekday := a.date.Weekday().String()
	abbreviatedWeekday := weekdayAbbreviations[fullWeekday]

	return fmt.Sprintf("%s, %s", abbreviatedWeekday, date)
}

func getTextLibraryPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
	}

	return fmt.Sprintf("%s/%s", home, relativeTextLibraryPath)
}

func getAssetLibraryPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
	}

	return fmt.Sprintf("%s/%s", home, relativeAssetLibraryPath)
}

func GetStartLocationQr() string {
	var loc string
	locations, err := getStartingLocations()
	if err != nil {
		panic(err)
	}

	loc, err = utils.FuzzyFind("Select starting Locations", locations)
	if err != nil {
		panic(err)
	}

	return loc
}

func getStartingLocations() ([]string, error) {
	var epsFiles []string

	dir := getTextLibraryPath() + "/meta/location-qr"

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
	switch a.date.Month() {
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
