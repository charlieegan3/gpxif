package operations

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCheckModTime(t *testing.T) {
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
