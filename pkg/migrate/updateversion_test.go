package migrate

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertOrUpdateVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		content string
	}{
		{
			`meta:
  version: v0`,
		}, {
			`meta:
  version: v3`,
		}, {
			``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()

			//nolint: gosec
			err := os.WriteFile(dir+"/header.yaml", []byte(tt.content), 0o644)
			if err != nil {
				t.Fatal(err)
			}

			res, err := InsertOrUpdateVersion(dir, "v3")
			if err != nil {
				t.Fatal(err)
			}

			assert.True(t, res)

			data, err := os.ReadFile(dir + "/header.yaml")
			if err != nil {
				t.Fatal(err)
			}

			// Check if the version was updated
			assert.Contains(t, string(data), "version: v3")
		})
	}
}
