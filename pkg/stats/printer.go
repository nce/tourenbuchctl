package stats

import (
	"os"

	md "github.com/nao1215/markdown"
)

func printMarkdown(activityCollection []activityData, regionalGrouping bool) {

	doc := md.NewMarkdown(os.Stdout)

	doc = doc.H1("MTB")

	regions := uniqueRegions(activityCollection)
	for _, region := range regions {
		var tableContent [][]string
		doc = doc.H2(region.Region)

		for _, activity := range activityCollection {
			if activity.Region == region.Region {
				line := []string{
					activity.Title, activity.Region, activity.Ascent, activity.Distance, activity.Duration,
				}
				tableContent = append(tableContent, line)
			}
		}

		doc = doc.Table(md.TableSet{
			Header: []string{"Name", "Region", "Ascent", "Distance", "Duration"},
			Rows:   tableContent,
		})

	}
	doc.Build()

}
