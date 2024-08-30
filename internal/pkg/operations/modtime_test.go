package operations

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckModTime(t *testing.T) {
	if os.Getenv("NIX_BUILD") == "true" {
		t.Skip()
	}

	testCases := map[string]struct {
		Image      string
		Operations []Operation
	}{
		"when update from UTC is needed": {
			Image: "../exif/fixtures/x100f.jpg",
			Operations: []Operation{
				{
					Reason:  "mtime != utc time",
					ModTime: true,
				},
			},
		},
		"when mtime is already set and no update is needed": {
			Image:      "../exif/fixtures/x100f-mtime.JPG",
			Operations: nil,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			operations, err := CheckModTime(testCase.Image)
			require.NoError(t, err)

			assert.Equal(t, testCase.Operations, operations)
		})
	}
}
