package main

import (
	"github.com/nce/tourenbuchctl/cmd"
)

var (
	version = "dev"
	//nolint: gochecknoglobals
	commit = "none"
	//nolint: gochecknoglobals
	date = "unknown"
)

func main() {
	cmd.Version = version
	cmd.Commit = commit
	cmd.Date = date
	cmd.Execute()
}
