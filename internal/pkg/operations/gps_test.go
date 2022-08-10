package operations

import (
	"github.com/charlieegan3/gpxif/internal/pkg/exif"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/charlieegan3/gpxif/internal/pkg/gpx"
)

func TestCheckGPSData(t *testing.T) {
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
					Reason:  "GPS data not found in EXIF",
					IFDPath: "IFD/GPSInfo",
					Fields: map[string]interface{}{
						"GPSLatitude":     exif.RationalDegreesMinutesSecondsFromDecimal(51.56734),
						"GPSLatitudeRef":  "N",
						"GPSLongitude":    exif.RationalDegreesMinutesSecondsFromDecimal(-0.13843),
						"GPSLongitudeRef": "W",
						"GPSAltitude":     []exifcommon.Rational{{Numerator: 75, Denominator: 1}},
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

			operations, err := CheckGPSData(testCase.Image, &g)
			require.NoError(t, err)

			assert.Equal(t, testCase.Operations, operations)
		})
	}
}
