package new

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/nce/tourenbuchctl/pkg/activity"
)

var (
	flags = &Flags{}
)

func NewNewCommand() *cobra.Command {
	var newCmd = &cobra.Command{
		Use:   "new",
		Short: "Create a new activity in Tourenbuch",
		Long:  "Create a new selected activity",
		Run: func(cmd *cobra.Command, args []string) {
			// Default action if no subcommands are specified
			fmt.Println("newnewnew")

		},
	}
	newCmd.AddCommand(newMtbCommand())
	newCmd.AddCommand(newSkitourCommand())

	return newCmd
}

func newMtbCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "mtb [name]",
		Short: "Create a new mtb activity in Tourenbuch",
		Long:  "Create a new mtb activity",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			flags.name = args[0]
			err := activity.CreateActivity(flags.name, flags.date, flags.rating)
			if err != nil {
				panic(err)
			}
		},
	}

	var dateStr string

	cmd.Flags().StringVarP(&dateStr, "date", "d", "", "Date of the activity in the format 'DD.MM.YYYY'")
	cmd.Flags().IntVarP(&flags.rating, "rating", "r", 3, "Rating of the activity in the format '1-5'."+
		"This will be later displayed as stars")

	err := cmd.MarkFlagRequired("date")
	if err != nil {
		panic(err)
	}

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if dateStr != "" {
			parsedDate, err := time.Parse("02.01.2006", dateStr)
			if err != nil {
				return fmt.Errorf("invalid date format: %v", err)
			}
			flags.date = parsedDate
		}
		return nil
	}

	return cmd

}

func newSkitourCommand() *cobra.Command {
	var newSkitourCmd = &cobra.Command{
		Use:   "skitour",
		Short: "Create a new skitour activity in Tourenbuch",
		Long:  "Create a new skitour activity",
		Run: func(cmd *cobra.Command, args []string) {
			// Default action if no subcommands are specified
			fmt.Println("Strava CLI application")
		},
	}
	return newSkitourCmd
}
