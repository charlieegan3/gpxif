package operations

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/charlieegan3/gpxif/internal/pkg/gpx"
)

func TTestCheckGPSData(t *testing.T) {
	testCases := map[string]struct {
		Image      string
		GPXFiles   []string
		Operations []Operation
	}{
		"when location is missing": {
			Image:    "../exif/fixtures/x100f.jpg",
			GPXFiles: []string{"./fixtures/2022-08-03.gpx"},
			Operations: []Operation{
				{
					Reason:  "Photo location is missing",
					IFDPath: "IFD/GPSInfo",
					Fields: map[string]string{
						"GPSLatitude":     "51/1 34/1 251/100",
						"GPSLatitudeRef":  "N",
						"GPSLongitudeRef": "W",
						"GPSLongitude":    "0/1 8/1 1936/100",
						"GPSAltitude":     "75",
					},
				},
			},
		},
		"when location is already set": {
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
