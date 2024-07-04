package activity

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
	"time"
)

// TODO:  make this configurable via external config file (-> viper)
const (
	relativeTextLibraryPath  = "vcs/github/nce/tourenbuch"
	relativeAssetLibraryPath = "Library/Mobile Documents/com~apple~CloudDocs/privat/sport/Tourenbuch/"
)

type Activity struct {
	category      string
	name          string
	textLocation  string
	assetLocation string
	date          time.Time
	rating        int
}

type ActivityClasses interface {
	CreateActivity(name string, date time.Time, rating int) error
}

// Each Tourenbuch entry is represented by two folders. One folder contains the
// text part of the entry, the other one contains the assets (images, etc.).
// Text is stored in git; Assets are stored in iCloud
func (a *Activity) createFolder() error {

	dirs := [2]string{
		getTextLibraryPath() + "/" + a.category + "/" + a.name + "-" + a.normalizeDate(),
		getAssetLibraryPath() + "/" + a.category + "/" + a.name + "-" + a.normalizeDate() + "/" + "img",
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
		Name  string
		Date  string
		Stars []struct{}
	}{
		Name:  a.name,
		Date:  a.normalizeDateWithShortWeekday(),
		Stars: make([]struct{}, a.rating),
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
