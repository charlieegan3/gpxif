package exif

import (
	"fmt"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestSetKey(t *testing.T) {
	testCases := map[string]struct {
		Image   string
		IFDPath string
		Key     string
		Value   any
	}{
		"set DateTimeOriginal": {
			Image:   "./fixtures/iphone.JPG",
			IFDPath: "IFD/Exif",
			Key:     "DateTimeOriginal",
			Value:   "2022:08:03 17:56:22",
		},
		"set OffsetTimeOriginal when missing in original": {
			Image:   "./fixtures/x100f.jpg",
			IFDPath: "IFD/Exif",
			Key:     "OffsetTimeOriginal",
			Value:   "+01:00",
		},
		"set GPSLatitude": {
			Image:   "./fixtures/iphone.JPG",
			IFDPath: "IFD/GPSInfo",
			Key:     "GPSLatitude",
			Value:   RationalDegreesMinutesSecondsFromDecimal(51.56736389),
		},
		//"set GPSLatitude when missing in original": {
		//	Image: "./fixtures/x100f.jpg",
		//	Key:   "GPSLatitude",
		//	Value: RationalDegreesMinutesSecondsFromDecimal(51.56736389),
		//},
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

				err = SetKey(imageCopy.Name(), testCase.IFDPath, testCase.Key, testCase.Value)
				require.NoError(t, err)

				readValue, err := GetKey(imageCopy.Name(), testCase.IFDPath, testCase.Key)
				require.NoError(t, err)

				assert.Equal(t, testCase.Value, readValue)
			}
		})
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

			newDateTime, err := GetKey(imageCopy.Name(), "IFD/Exif", "DateTimeOriginal")
			require.NoError(t, err)
			newOffset, err := GetKey(imageCopy.Name(), "IFD/Exif", "OffsetTimeOriginal")
			require.NoError(t, err)
			newSubSecTime, err := GetKey(imageCopy.Name(), "IFD/Exif", "SubSecTimeOriginal")
			require.NoError(t, err)

			assert.Equal(t, testCase.LocalTime.Format("2006-01-02 15:04:05"), newDateTime)
			assert.Equal(t, testCase.LocalTime.Format("-07:00"), newOffset)
			assert.Equal(t, fmt.Sprintf("%d", testCase.LocalTime.Nanosecond()/1000000), newSubSecTime)
		})
	}
}

func TestGetKey(t *testing.T) {
	testCases := map[string]struct {
		Image         string
		IFDPath       string
		Key           string
		ExpectedValue interface{}
	}{
		"get DateTimeOriginal": {
			Image:         "./fixtures/iphone.JPG",
			IFDPath:       "IFD/Exif",
			Key:           "DateTimeOriginal",
			ExpectedValue: "2022:08:03 18:56:22",
		},
		"get OffsetTimeOriginal": {
			Image:         "./fixtures/iphone.JPG",
			IFDPath:       "IFD/Exif",
			Key:           "OffsetTimeOriginal",
			ExpectedValue: "+01:00",
		},
		"get SubsecTimeOriginal": {
			Image:         "./fixtures/iphone.JPG",
			IFDPath:       "IFD/Exif",
			Key:           "SubSecTimeOriginal",
			ExpectedValue: "480",
		},
		"get GPSLatitude": {
			Image:   "./fixtures/iphone.JPG",
			IFDPath: "IFD/GPSInfo",
			Key:     "GPSLatitude",
			ExpectedValue: []exifcommon.Rational{
				{51, 1},
				{34, 1},
				{251, 100},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			value, err := GetKey(testCase.Image, testCase.IFDPath, testCase.Key)
			require.NoError(t, err)

			assert.Equal(t, testCase.ExpectedValue, value)
		})
	}
}

func TestRationalDegreesMinutesSecondsFromDecimal(t *testing.T) {
	testCases := map[string]struct {
		Decimal       float64
		ExpectedValue []exifcommon.Rational
	}{
		"example 1": {
			Decimal: 53.383328,
			ExpectedValue: []exifcommon.Rational{
				{53, 1},
				{22, 1},
				{5998, 100},
			},
		},
		"example 2": {
			Decimal: 75.1234,
			ExpectedValue: []exifcommon.Rational{
				{75, 1},
				{7, 1},
				{2424, 100},
			},
		},
		"example 3": {
			Decimal: -0.13843,
			ExpectedValue: []exifcommon.Rational{
				{0, 1},
				{8, 1},
				{1834, 100},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.ExpectedValue, RationalDegreesMinutesSecondsFromDecimal(testCase.Decimal))
		})
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
	}
}
