package operations

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/charlieegan3/gpxif/internal/pkg/gpx"
)

func TestCheckLocalTime(t *testing.T) {
	testCases := map[string]struct {
		Image      string
		GPXFiles   []string
		Operations []Operation
	}{
		"when update from UTC is needed": {
			Image:    "../exif/fixtures/x100f.jpg",
			GPXFiles: []string{"./fixtures/2022-08-03.gpx"},
			Operations: []Operation{
				{
					Reason:  "DateTimeOriginal data was not in local time",
					IFDPath: "IFD/Exif",
					Fields: map[string]string{
						"DateTimeOriginal":   "2022:08:03 18:57:55",
						"OffsetTimeOriginal": "+01:00",
						"SubSecTimeOriginal": "0",
					},
				},
			},
		},
		"when local time is already set and no update is needed": {
			Image:      "../exif/fixtures/iphone.JPG",
			GPXFiles:   []string{"./fixtures/2022-08-03.gpx"},
			Operations: nil,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			g, err := gpx.NewGPXDatasetFromFile(testCase.GPXFiles...)
			require.NoError(t, err)

			operations, err := CheckLocalTime(testCase.Image, g)
			require.NoError(t, err)

			assert.Equal(t, testCase.Operations, operations)
		})
	}
}
