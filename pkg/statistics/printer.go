package statistics

import (
	"os"
	"sort"

	md "github.com/nao1215/markdown"
	"github.com/rs/zerolog/log"
)

func printMarkdown(activityCollection []activityData, regionalGrouping bool) {
	doc := md.NewMarkdown(os.Stdout)

	if regionalGrouping {
		regions := uniqueRegions(activityCollection)
		for _, region := range regions {
			doc = doc.H2(region.Region)

			var tableContent [][]string

			for _, activity := range activityCollection {
				if activity.Region == region.Region {
					line := []string{
						"[" + activity.Title + "](" + activity.Dirname + ")" + " (" + activity.Date + ")",
						"[📚](https://s3.nce.wtf/" + activity.Type + "/" + activity.Dirname + ".pdf)",
						activity.Ascent + "hm",
						activity.Distance + "km",
						activity.Duration + "h",
						activity.Participants,
					}
					tableContent = append(tableContent, line)
				}
			}

			sort.Slice(tableContent, func(i, j int) bool {
				return tableContent[i][0] < tableContent[j][0]
			})

			doc = doc.Table(md.TableSet{
				Header: []string{"Name", "PDF", "Ascent", "Distance", "Duration", "Participants"},
				Rows:   tableContent,
			})
		}
	} else {
		var tableContent [][]string

		for _, activity := range activityCollection {
			line := []string{
				"[" + activity.Title + "](" + activity.Dirname + ")" + " (" + activity.Date + ")",
				activity.Ascent + "hm",
				activity.Distance + "km",
				activity.Duration + "h",
				activity.Participants,
			}
			tableContent = append(tableContent, line)
		}

		sort.Slice(tableContent, func(i, j int) bool {
			return tableContent[i][0] < tableContent[j][0]
		})

		doc = doc.Table(md.TableSet{
			Header: []string{"Name", "PDF", "Ascent", "Distance", "Duration", "Participants"},
			Rows:   tableContent,
		})
	}

	err := doc.Build()
	if err != nil {
		log.Error().Msgf("Error building markdown output: %v", err)
	}
}
