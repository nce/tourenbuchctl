package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tourenbuch",
	Short: "tourenbuch CLI application",
	Long:  "A CLI application to interact with the Strava API.",
	Run: func(cmd *cobra.Command, args []string) {
		// Default action if no subcommands are specified
		fmt.Println("Strava CLI application")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
