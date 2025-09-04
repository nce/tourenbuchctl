package statistics

import "errors"

var (
	ErrTourenbuchDirNameWrong = errors.New("directory name does not match expected schema")
	ErrNoValidActivityTypes   = errors.New("no valid activity types found")
)
