package gpx

import (
	"os"
	"testing"
	"time"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkrajina/gpxgo/gpx"
)

func TestNewGPXDatasetFromReader(t *testing.T) {
	testCases := map[string]struct {
		File                   string
		ExpectedFirstPointTime time.Time
		ExpectedFirstPoint     gpx.Point
		ExpectedLastPointTime  time.Time
		ExpectedLastPoint      gpx.Point
		ExpectPointCount       int
	}{
		"simple example": {
			File:                   "fixtures/run.gpx",
			ExpectedFirstPointTime: time.Date(2022, time.August, 3, 8, 10, 0, 0, time.UTC),
			ExpectedFirstPoint: gpx.Point{
				Latitude:  51.5671980,
				Longitude: -0.1413280,
				Elevation: *gpx.NewNullableFloat64(90.2),
			},
			ExpectedLastPointTime: time.Date(2022, time.August, 3, 8, 10, 0, 0, time.UTC),
			ExpectedLastPoint: gpx.Point{
				Latitude:  51.5671980,
				Longitude: -0.1413280,
				Elevation: *gpx.NewNullableFloat64(90.2),
			},
			ExpectPointCount: 10,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			file, err := os.Open(testCase.File)
			require.NoError(t, err)

			gpxDataset, err := NewGPXDatasetFromReader(file)
			require.NoError(t, err)

			firstPoint, err := gpxDataset.AtTime(testCase.ExpectedFirstPointTime)
			require.NoError(t, err)

			lastPoint, err := gpxDataset.AtTime(testCase.ExpectedLastPointTime)
			require.NoError(t, err)

			assert.Equal(t, testCase.ExpectedFirstPoint, firstPoint.Point)
			assert.Equal(t, testCase.ExpectedLastPoint, lastPoint.Point)

		})
	}
}

func TestInRange(t *testing.T) {
	testCases := map[string]struct {
		Files          []string
		Time           time.Time
		ExpectedResult bool
	}{
		"when in range": {
			Files:          []string{"fixtures/run.gpx", "fixtures/run_nida.gpx"},
			Time:           time.Date(2022, time.August, 3, 8, 10, 0, 0, time.UTC),
			ExpectedResult: true,
		},
		"run example with no match": {
			Files:          []string{"fixtures/run.gpx", "fixtures/run_nida.gpx"},
			Time:           time.Date(2022, time.July, 27, 8, 10, 0, 0, time.UTC),
			ExpectedResult: false,
		},
		"when no data": {
			Files:          []string{},
			Time:           time.Date(2022, time.August, 3, 8, 10, 0, 0, time.UTC),
			ExpectedResult: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			gpxDataset, err := NewGPXDatasetFromDisk(testCase.Files...)
			require.NoError(t, err)

			result, _ := gpxDataset.InRange(testCase.Time)

			assert.Equal(t, testCase.ExpectedResult, result)
		})
	}
}

func TestAtTime(t *testing.T) {
	testCases := map[string]struct {
		File          string
		Time          time.Time
		ExpectedPoint *gpx.GPXPoint
		ExpectedError *string
	}{
		"run example with match": {
			File: "fixtures/run.gpx",
			Time: time.Date(2022, time.August, 3, 8, 10, 0, 0, time.UTC),
			ExpectedPoint: &gpx.GPXPoint{
				Point:     gpx.Point{Latitude: 51.5671980, Longitude: -0.1413280},
				Timestamp: time.Date(2022, time.August, 3, 8, 9, 57, 0, time.UTC),
			},
		},
		"run example with no match returns last/first point": {
			File: "fixtures/run.gpx",
			Time: time.Date(2022, time.August, 3, 8, 53, 0, 0, time.UTC),
			ExpectedPoint: &gpx.GPXPoint{
				Point:     gpx.Point{Latitude: 51.5673220, Longitude: -0.1383400},
				Timestamp: time.Date(2022, time.August, 3, 8, 52, 9, 0, time.UTC),
			},
		},
		"run example with no match and +24h offset returns error": {
			File:          "fixtures/run.gpx",
			Time:          time.Date(2022, time.August, 4, 8, 53, 0, 0, time.UTC),
			ExpectedError: strPtr("out of range of loaded files"),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			gpxDataset, err := NewGPXDatasetFromDisk(testCase.File)
			require.NoError(t, err)

			point, err := gpxDataset.AtTime(testCase.Time)

			if testCase.ExpectedError != nil {
				require.ErrorContains(t, err, *testCase.ExpectedError)
			} else {
				require.NoError(t, err)
			}

			if testCase.ExpectedPoint != nil {
				expectedPoint := td.SStruct(
					testCase.ExpectedPoint,
					td.StructFields{
						"=*": td.Ignore(),
					})

				td.Cmp(t, &point, expectedPoint)
			}
		})
	}
}

func strPtr(str string) *string {
	return &str
}
