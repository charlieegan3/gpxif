package exif

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGetUTC(t *testing.T) {
	testCases := map[string]struct {
		Image           string
		ExpectedUTCTime time.Time
	}{
		"when iphone": {
			Image:           "./fixtures/iphone.JPG",
			ExpectedUTCTime: time.Date(2022, time.August, 3, 17, 56, 22, 0, time.UTC),
		},
		"when iphone in another timezone": {
			Image:           "./fixtures/iphone_other_tz.JPG",
			ExpectedUTCTime: time.Date(2022, time.July, 30, 17, 57, 04, 0, time.UTC),
		},
		"when iphone with no offset": {
			Image:           "./fixtures/iphone_no_offset.JPG",
			ExpectedUTCTime: time.Date(2022, time.January, 21, 9, 9, 0, 0, time.UTC),
		},
		"when iphone moment DNG": {
			Image:           "./fixtures/moment.DNG",
			ExpectedUTCTime: time.Date(2022, time.August, 3, 17, 55, 54, 0, time.UTC),
		},
		"when iphone HIEC": {
			Image:           "./fixtures/iphone.HEIC",
			ExpectedUTCTime: time.Date(2022, time.August, 3, 17, 57, 45, 0, time.UTC),
		},
		"when x100f jpg with no offset": {
			Image:           "./fixtures/x100f.jpg",
			ExpectedUTCTime: time.Date(2022, time.August, 3, 17, 57, 55, 0, time.UTC),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			time, err := GetUTC(testCase.Image)
			require.NoError(t, err)

			assert.Equal(t, testCase.ExpectedUTCTime, time)
		})
		continue
	}
}
