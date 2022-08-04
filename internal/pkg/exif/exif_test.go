package exif

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestSetKeyString(t *testing.T) {
	testCases := map[string]struct {
		Image string
		Key   string
		Value string
	}{
		"set DateTimeOriginal": {
			Image: "./fixtures/iphone.JPG",
			Key:   "DateTimeOriginal",
			Value: "2022:08:03 17:56:22",
		},
		"set OffsetTimeOriginal when missing in original": {
			Image: "./fixtures/x100f.jpg",
			Key:   "OffsetTimeOriginal",
			Value: "+01:00",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			imageCopy, err := ioutil.TempFile(".", "image_")
			require.NoError(t, err)
			defer os.Remove(imageCopy.Name())

			imageFile, err := os.Open(testCase.Image)
			require.NoError(t, err)

			_, err = io.Copy(imageCopy, imageFile)
			require.NoError(t, err)

			// test that we can mutate the same file again
			count := 0
			for {
				if count > 1 {
					break
				}
				count++

				err = SetKeyString(imageCopy.Name(), testCase.Key, testCase.Value)
				require.NoError(t, err)

				readValue, err := GetKeyString(imageCopy.Name(), testCase.Key)
				require.NoError(t, err)

				assert.Equal(t, testCase.Value, readValue)
			}
		})
		continue
	}
}

func TestSetLocalTime(t *testing.T) {
	location, err := time.LoadLocation("Europe/London")
	require.NoError(t, err)

	testCases := map[string]struct {
		Image     string
		Key       string
		LocalTime time.Time
	}{
		"set DateTimeOriginal to local time": {
			Image:     "./fixtures/x100f.jpg",
			Key:       "DateTimeOriginal",
			LocalTime: time.Date(2022, time.July, 31, 20, 13, 21, 500000000, location),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			// make a copy of the image
			imageCopy, err := ioutil.TempFile(".", "image_")
			require.NoError(t, err)
			defer os.Remove(imageCopy.Name())

			imageFile, err := os.Open(testCase.Image)
			require.NoError(t, err)

			_, err = io.Copy(imageCopy, imageFile)
			require.NoError(t, err)

			// set the local time for the image and check it's set correctly
			err = SetLocalTime(imageCopy.Name(), testCase.LocalTime)
			require.NoError(t, err)

			newDateTime, err := GetKeyString(imageCopy.Name(), "DateTimeOriginal")
			require.NoError(t, err)
			newOffset, err := GetKeyString(imageCopy.Name(), "OffsetTimeOriginal")
			require.NoError(t, err)
			newSubSecTime, err := GetKeyString(imageCopy.Name(), "SubSecTimeOriginal")
			require.NoError(t, err)

			assert.Equal(t, testCase.LocalTime.Format("2006-01-02 15:04:05"), newDateTime)
			assert.Equal(t, testCase.LocalTime.Format("-07:00"), newOffset)
			assert.Equal(t, fmt.Sprintf("%d", testCase.LocalTime.Nanosecond()/1000000), newSubSecTime)
		})
		continue
	}
}

func TestGetKeyString(t *testing.T) {
	testCases := map[string]struct {
		Image         string
		Key           string
		ExpectedValue string
	}{
		"get DateTimeOriginal": {
			Image:         "./fixtures/iphone.JPG",
			Key:           "DateTimeOriginal",
			ExpectedValue: "2022:08:03 18:56:22",
		},
		"get OffsetTimeOriginal": {
			Image:         "./fixtures/iphone.JPG",
			Key:           "OffsetTimeOriginal",
			ExpectedValue: "+01:00",
		},
		"get SubsecTimeOriginal": {
			Image:         "./fixtures/iphone.JPG",
			Key:           "SubSecTimeOriginal",
			ExpectedValue: "480",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			value, err := GetKeyString(testCase.Image, testCase.Key)
			require.NoError(t, err)

			assert.Equal(t, testCase.ExpectedValue, value)
		})
		continue
	}
}

func TestGetUTC(t *testing.T) {
	testCases := map[string]struct {
		Image           string
		ExpectedUTCTime time.Time
	}{
		"when iphone": {
			Image:           "./fixtures/iphone.JPG",
			ExpectedUTCTime: time.Date(2022, time.August, 3, 17, 56, 22, 480000000, time.UTC),
		},
		"when iphone in another timezone": {
			Image:           "./fixtures/iphone_other_tz.JPG",
			ExpectedUTCTime: time.Date(2022, time.July, 30, 17, 57, 04, 349000000, time.UTC),
		},
		"when iphone with no offset": {
			Image:           "./fixtures/iphone_no_offset.JPG",
			ExpectedUTCTime: time.Date(2022, time.January, 21, 9, 9, 0, 97000000, time.UTC),
		},
		"when iphone moment DNG": {
			Image:           "./fixtures/moment.DNG",
			ExpectedUTCTime: time.Date(2022, time.August, 3, 17, 55, 54, 222000000, time.UTC),
		},
		"when iphone HIEC": {
			Image:           "./fixtures/iphone.HEIC",
			ExpectedUTCTime: time.Date(2022, time.August, 3, 17, 57, 45, 986000000, time.UTC),
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
