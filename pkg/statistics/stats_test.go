package statistics

import (
	"reflect"
	"testing"

	"github.com/nce/tourenbuchctl/pkg/activity"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func setupApp(t *testing.T) {
	t.Helper()

	viper.Set("activities", []string{"mtb", "skitour"})
	activity.SetupActivityKinds()
}

func TestFilterActivityKinds(t *testing.T) {
	setupApp(t)
	t.Parallel()

	tests := []struct {
		activityKinds string
		expected      []activity.Kind
	}{
		{
			"all",
			[]activity.Kind{
				{Name: "mtb"},
				{Name: "skitour"},
			},
		}, {
			"mtb",
			[]activity.Kind{
				{Name: "mtb"},
			},
		}, {
			"mtb, skitour",
			[]activity.Kind{
				{Name: "mtb"},
				{Name: "skitour"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.activityKinds, func(t *testing.T) {
			t.Parallel()

			result, err := filterActivityKinds(tt.activityKinds)
			if !assert.Equal(t, reflect.TypeOf([]activity.Kind{}), reflect.TypeOf(result)) {
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

func TestFilterActivityKindsForErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		activityKinds string
		expected      []activity.Kind
	}{
		{
			"notvalid",
			[]activity.Kind{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.activityKinds, func(t *testing.T) {
			t.Parallel()

			_, err := filterActivityKinds(tt.activityKinds)
			if err == nil {
				t.Error("expected an error")
			}
		})
	}
}
