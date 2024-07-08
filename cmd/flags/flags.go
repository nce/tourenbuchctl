package flags

import "time"

type CreateFlags struct {
	Name               string
	Title              string
	Date               time.Time
	StravaSync         bool
	QueryStartLocation bool
	StartLocation      string
	StartLocationQr    string
}

type CreateMtbFlags struct {
	Core       CreateFlags
	Rating     int
	Company    string
	Restaurant string
	Difficulty int
}
