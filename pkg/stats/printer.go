package stats

import (
	"os"
	"sort"

	md "github.com/nao1215/markdown"
	"github.com/rs/zerolog/log"
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

		sort.Slice(tableContent, func(i, j int) bool {
			return tableContent[i][0] < tableContent[j][0]
		})

		doc = doc.Table(md.TableSet{
			Header: []string{"Name", "Region", "Ascent", "Distance", "Duration"},
			Rows:   tableContent,
		})
	}

	err := doc.Build()
	if err != nil {
		log.Error().Msgf("Error building markdown output: %v", err)
	}
}
