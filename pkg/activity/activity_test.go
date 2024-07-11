package activity

import (
	"fmt"
	"testing"
	"time"
)

func TestNormalizeStartTime(t *testing.T) {
	tests := []struct {
		start    time.Time
		expected string
	}{
		{time.Date(2021, 8, 15, 14, 30, 45, 100, time.Local), "14:30"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			start := Activity{
				Tb: Tourenbuch{
					StartTime: tt.start,
				},
			}
			result := start.normalizeStartTime()
			if result != tt.expected {
				t.Errorf("got %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestNormalizeDistance(t *testing.T) {
	tests := []struct {
		meters   int
		expected string
	}{
		{3661, "3.7"},
		{0, "0.0"},
		{999, "1.0"},
		{1010, "1.0"},
		{60000, "60.0"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d meters", tt.meters), func(t *testing.T) {
			distance := Activity{
				Tb: Tourenbuch{
					Distance: tt.meters,
				},
			}
			result := distance.normalizeDistance()
			if result != tt.expected {
				t.Errorf("got %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestNormalizeAscent(t *testing.T) {
	tests := []struct {
		meters   int
		expected string
	}{
		{3661, "3.661"},
		{0, "0"},
		{999, "999"},
		{60000, "60.000"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d meters", tt.meters), func(t *testing.T) {
			distance := Activity{
				Tb: Tourenbuch{
					Ascent: tt.meters,
				},
			}
			result := distance.normalizeAscent()
			if result != tt.expected {
				t.Errorf("got %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestNormalizeDuration(t *testing.T) {
	tests := []struct {
		seconds  int
		expected string
	}{
		{3661, "01:01"},
		{0, "00:00"},
		{59, "00:00"},
		{60, "00:01"},
		{3600, "01:00"},
		{86399, "23:59"},
		{86400, "24:00"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d seconds", tt.seconds), func(t *testing.T) {
			result := normalizeDuration(time.Duration(tt.seconds) * time.Second)
			if result != tt.expected {
				t.Errorf("got %s, want %s", result, tt.expected)
			}
		})
	}
}
