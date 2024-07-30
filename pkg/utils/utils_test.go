package utils

import "testing"

func TestSplitDirectoryName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		dir  string
		name string
		date string
	}{
		{"3.laender-01.02.2001", "3.laender", "01.02.2001"},
	}

	for _, tt := range tests {
		t.Run(tt.dir+" dirs", func(t *testing.T) {
			t.Parallel()

			name, date, _ := SplitActivityDirectoryName(tt.dir)

			if name != tt.name {
				t.Errorf("got %s, want %s", name, tt.name)
			}

			if date != tt.date {
				t.Errorf("got %s, want %s", date, tt.date)
			}
		})
	}
}
