package stats

import (
	"reflect"
	"testing"

	"github.com/nce/tourenbuchctl/pkg/activity"
	"github.com/stretchr/testify/assert"
)

func TestFilterActivityTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		activityTypes string
		expected      []activity.ActivityType
	}{
		{
			"all",
			[]activity.ActivityType{
				{Name: "mtb"},
				{Name: "skitour"},
			},
		}, {
			"mtb",
			[]activity.ActivityType{
				{Name: "mtb"},
			},
		}, {
			"mtb, skitour",
			[]activity.ActivityType{
				{Name: "mtb"},
				{Name: "skitour"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.activityTypes, func(t *testing.T) {
			t.Parallel()

			result, err := filterActivityTypes(tt.activityTypes)
			if !assert.Equal(t, reflect.TypeOf([]activity.ActivityType{}), reflect.TypeOf(result)) {
				t.Errorf("got %v, want %v", result, tt.expected)
			}

			if len(tt.expected) != len(result) {
				t.Errorf("got %d, want %d", len(result), len(tt.expected))
			}

			if err != nil {
				t.Errorf("got %v", err)
			}
		})
	}
}

func TestFilterActivityTypesForErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		activityTypes string
		expected      []activity.ActivityType
	}{
		{
			"notvalid",
			[]activity.ActivityType{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.activityTypes, func(t *testing.T) {
			t.Parallel()

			_, err := filterActivityTypes(tt.activityTypes)
			if err == nil {
				t.Error("expected an error")
			}
		})
	}
}
