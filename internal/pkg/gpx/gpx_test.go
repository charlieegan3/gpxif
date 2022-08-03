package gpx

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/require"
	"github.com/tkrajina/gpxgo/gpx"
)

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
			gpxDataset, err := NewGPXDatasetFromFile(testCase.Files...)
			require.NoError(t, err)

			result := gpxDataset.InRange(testCase.Time)

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
		"run example with no match": {
			File:          "fixtures/run.gpx",
			Time:          time.Date(2021, time.August, 3, 8, 10, 0, 0, time.UTC),
			ExpectedError: strPtr("out of range of loaded files"),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			gpxDataset, err := NewGPXDatasetFromFile(testCase.File)
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
