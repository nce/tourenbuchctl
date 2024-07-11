package activity

import (
	"os"

	"github.com/nce/tourenbuchctl/pkg/strava"
)

func (a *Activity) CreateActivity() error {

	err := a.createFolder()
	if err != nil {
		return err
	}

	for _, file := range []string{"description.md", "elevation.plt", "Makefile", "img-even.tex"} {

		text, err := a.initSkeleton(file)
		if err != nil {
			return err
		}

		file, err := os.Create(a.Meta.TextLocation + "/" + file)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = file.WriteString(text)
		if err != nil {
			return err
		}
	}

	if a.Meta.StravaSync {
		stats, err := strava.FetchStravaData(a.Tb.Date)

		if err != nil {
			return err
		}

		a.Tb.Distance = stats.Distance
		a.Tb.Ascent = stats.Ascent
		a.Tb.StartTime = stats.StartDate
		a.Tb.ElapsedTime = stats.ElapsedTime
		a.Tb.MovingTime = stats.MovingTime

		err = a.updateActivity(a.Meta.TextLocation + "/description.md")
		if err != nil {
			return err
		}
	}

	return nil
}
