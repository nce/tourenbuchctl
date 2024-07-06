package activity

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

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
	}

	io := new(strings.Builder)

	// Execute the template and write to the file
	err = tmpl.Execute(io, data)
	if err != nil {
		panic(err)
	}
	return io.String(), nil
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

	dir := getAssetLibraryPath() + "/meta/location-qr"

	// Read directory contents
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// Filter for .eps files
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".eps" {
			fileName := strings.TrimSuffix(file.Name(), ".eps")
			epsFiles = append(epsFiles, fileName)
		}
	}

	return epsFiles, nil
}
