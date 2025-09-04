package activity

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// nolint: gochecknoglobals
var ActivityTypes []Kind

type Kind struct {
	Name      string
	TextPath  string
	AssetPath string
}

func GetActivityKinds() {
	validActivities := viper.GetStringSlice("activities")

	textPath, err := GetTextLibraryPath()
	if err != nil {
		log.Error().Msg("Error getting Library Path")
	}

	assetPath, err := GetAssetLibraryPath()
	if err != nil {
		log.Error().Msg("Error getting Asset Path")
	}

	for _, validActivity := range validActivities {
		activity := Kind{
			Name:      validActivity,
			TextPath:  textPath + "/" + validActivity,
			AssetPath: assetPath + "/" + validActivity,
		}

		ActivityTypes = append(ActivityTypes, activity)
	}
}
