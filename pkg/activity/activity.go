package activity

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/nce/tourenbuchctl/pkg/migrate"
	"github.com/nce/tourenbuchctl/pkg/utils"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v3"
)

// make this configurable via external config file (-> viper).
const (
	relativeTextLibraryPath  = "vcs/github/nce/tourenbuch/"
	relativeAssetLibraryPath = "Library/Mobile Documents/com~apple~CloudDocs/privat/sport/Tourenbuch/"
)

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
// Text is stored in git; Assets are stored in iCloud.
func (a *Activity) createFolder() error {
	textPath, err := GetTextLibraryPath()
	if err != nil {
		return err
	}

	assetPath, err := GetAssetLibraryPath()
	if err != nil {
		return err
	}

	// this is duplicated and should be refactored
	var dirs [2]string
	if a.Meta.Multiday {
		dirs = [2]string{
			textPath + a.Meta.Category + "/" + "multidaytrip/" + a.Meta.Name,
			assetPath + a.Meta.Category + "/" + "multidaytrip/" + a.Meta.Name + "/" + "img",
		}
	} else {
		dirs = [2]string{
			textPath + a.Meta.Category + "/" + a.Meta.Name + "-" + a.normalizeDate(),
			assetPath + a.Meta.Category + "/" + a.Meta.Name + "-" + a.normalizeDate() + "/" + "img",
		}
	}

	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0o755)
		if err != nil {
			return fmt.Errorf("creating new Asset/Lib Dir: %s; error: %w", dir, err)
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
		return "", fmt.Errorf("parsing template file: %s; error: %w", location, err)
	}

	textPath, err := GetTextLibraryPath()
	if err != nil {
		return "", fmt.Errorf("getting Library Path %w", err)
	}

	assetPath, err := GetAssetLibraryPath()
	if err != nil {
		return "", fmt.Errorf("getting Asset Path %w", err)
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

	newString := new(strings.Builder)

	// Execute the template and write to the file
	err = tmpl.Execute(newString, data)
	if err != nil {
		return "", fmt.Errorf("executing TB template: %w", err)
	}

	return newString.String(), nil
}

// Updates the existing stats-yaml structure with new data.
//
//gocyclo:ignore
func (a *Activity) updateActivity(dir string) error {
	migrated, err := migrate.SplitDescriptionFile(a.Meta.TextLocation)
	if err != nil {
		return fmt.Errorf("migration error in %s: %w", a.Meta.TextLocation, err)
	}

	if migrated {
		log.Info().Str("filename", a.Meta.TextLocation).
			Msg("Description split into header.yaml and description.md")
	}

	file := dir + "header.yaml"

	yamlFile, err := os.ReadFile(file)
	if err != nil {
		log.Error().Str("filename", file).Msg("Error reading file")

		return fmt.Errorf("reading file %w", err)
	}

	// Parse the YAML file into a node tree
	var root yaml.Node

	err = yaml.Unmarshal(yamlFile, &root)
	if err != nil {
		log.Error().Str("filename", file).Msg("Error unmarshalling file")

		return fmt.Errorf("unmarshalling file %w", err)
	}

	// No idea. This is written by AI
	// This modifies just one nested Yaml Node, without touching/killing the
	// rest of the file. It updates the statistic data (distance/ascent) of the activity

	// Navigate to the "stats" node and modify it
	// Traverse the document to find the "stats" key
	for i := range len(root.Content) {
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
		log.Error().Str("filename", file).Msg("Serializing to yaml failed")

		return fmt.Errorf("serialzing yaml %w", err)
	}

	// Write the modified YAML string back to the file
	//nolint: gosec
	err = os.WriteFile(file, output, 0o644)
	if err != nil {
		log.Error().Str("filename", file).Msg("Writing back to file failed")

		return fmt.Errorf("writing to file %w", err)
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
	loc, _ := time.LoadLocation("Europe/Berlin")
	localTime := a.Tb.StartTime.In(loc)

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
		return "", "", err
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
		return "", fmt.Errorf("fuzzy finding starting locatin: %w", err)
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
