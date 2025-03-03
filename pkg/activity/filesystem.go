package activity

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

func (a *Activity) CreateActivity() error {
	err := a.createFolder()
	if err != nil {
		return fmt.Errorf("error creating folder: %w", err)
	}

	for _, file := range []string{"description.md", "header.yaml", "elevation.plt", "images.tex"} {
		text, err := a.initSkeleton(file)
		if err != nil {
			return fmt.Errorf("creating init skeleton: %w", err)
		}

		file, err := os.Create(a.Meta.TextLocation + "/" + file)
		if err != nil {
			return fmt.Errorf("error creating file: %w", err)
		}
		defer file.Close()

		_, err = file.WriteString(text)
		if err != nil {
			return fmt.Errorf("error writing file: %w", err)
		}
	}

	err = a.StravaSync()
	if err != nil {
		return fmt.Errorf("error syncing new activity with strava: %w", err)
	}

	return nil
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
	location := fmt.Sprintf("templates/%s/%s", a.Meta.Category, file)

	tmpl, err := template.ParseFS(content, location)
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
		TrailDifficulty  int
		SkiDifficulty    string
		AvalancheReport  int
		MaxElevation     string
		Restaurant       string
		Season           string
		Runs             int
		Vertical         string
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
		TrailDifficulty:  a.Tb.TrailDifficulty,
		SkiDifficulty:    a.Tb.SkiDifficulty,
		AvalancheReport:  a.Tb.AvalancheReport,
		MaxElevation:     normalizeElevation(a.Tb.MaxElevation),
		Restaurant:       a.Tb.Restaurant,
		Season:           a.getSeason(),
		Runs:             a.Tb.AlpineSki.Runs,
		Vertical:         normalizeElevation(a.Tb.AlpineSki.Vertical),
	}

	newString := new(strings.Builder)

	// Execute the template and write to the file
	err = tmpl.Execute(newString, data)
	if err != nil {
		return "", fmt.Errorf("executing TB template: %w", err)
	}

	return newString.String(), nil
}
