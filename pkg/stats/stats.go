package stats

import (
	"fmt"
	"strings"

	"github.com/nce/tourenbuchctl/pkg/activity"
)

// filter activity types by string inputs like "mtb" or "mtb, skitour"
func filterActivityTypes(activityTypes string) ([]activity.ActivityType, error) {
	var validActivityTypes []activity.ActivityType

	if activityTypes == "all" {

		for _, activityType := range activity.ActivityTypes {
			validActivityTypes = append(validActivityTypes, activityType)
		}
		return validActivityTypes, nil
	}

	unfilteredActivityTypes := strings.Split(strings.ReplaceAll(activityTypes, " ", ""), ",")
	var filteredActivityTypes []activity.ActivityType

	for _, unfilteredActivityType := range unfilteredActivityTypes {
		// check if it's a valid activity type
		if activity.ValidActivityType(unfilteredActivityType) {
			for _, validType := range activity.ActivityTypes {
				// add this valid type to the slice
				if validType.Name == unfilteredActivityType {
					filteredActivityTypes = append(filteredActivityTypes, validType)
				}
			}
		}
	}

	if len(filteredActivityTypes) == 0 {
		return nil, fmt.Errorf("no valid activity types found")
	}

	return filteredActivityTypes, nil
}

func gatherDirectories() {}
