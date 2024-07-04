package cmd

import (
	"fmt"
	"os"

	"github.com/nce/tourenbuchctl/cmd/new"
	"github.com/nce/tourenbuchctl/cmd/sync"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tourenbuchctl",
	Short: "tourenbuch CLI application",
	Long:  "A CLI application to interact with Tourenbuch.",
	Run: func(cmd *cobra.Command, args []string) {
		// Default action if no subcommands are specified
		fmt.Println("Tourenbuch CLI application")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(sync.NewSyncCommand())
	rootCmd.AddCommand(new.NewNewCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
