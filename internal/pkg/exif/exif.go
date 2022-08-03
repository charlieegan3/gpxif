package exif

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
)

func GetUTC(image string) (time.Time, error) {
	b, err := ioutil.ReadFile(image)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read image file: %w", err)
	}

	rawExif, err := exif.SearchAndExtractExif(b)
	if err == exif.ErrNoExif {
		return time.Time{}, fmt.Errorf("no exif data found")
	} else if err != nil {
		return time.Time{}, fmt.Errorf("failed to get raw exif data: %s", err)
	}

	im, err := exifcommon.NewIfdMappingWithStandard()
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to create idfmapping: %s", err)
	}

	ti := exif.NewTagIndex()

	_, index, err := exif.Collect(im, ti, rawExif)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to collect exif data: %s", err)
	}

	var dateTimeOriginal time.Time
	var dateTimeOriginalSubSec int
	var offset time.Time

	cb := func(ifd *exif.Ifd, ite *exif.IfdTagEntry) error {
		if ite.TagName() == "DateTimeOriginal" {
			rawValue, err := ite.Value()
			if err != nil {
				return fmt.Errorf("could not get raw DateTimeOriginal value")
			}

			val, ok := rawValue.(string)
			if !ok {
				return fmt.Errorf("DateTimeOriginal was not in expected format: %#v", rawValue)
			}

			dateTimeOriginal, err = time.Parse("2006:01:02 15:04:05", string(val))
			if err != nil {
				return fmt.Errorf("failed to parse DateTimeOriginal: %s", err)
			}
		}
		if ite.TagName() == "SubsecTimeOriginal" {
			rawValue, err := ite.Value()
			if err != nil {
				return fmt.Errorf("could not get raw SubsecTimeOriginal value")
			}

			val, ok := rawValue.(int)
			if !ok {
				return fmt.Errorf("SubsecTimeOriginal was not in expected format: %#v", rawValue)
			}

			dateTimeOriginalSubSec = val
		}

		if ite.TagName() == "OffsetTimeOriginal" {
			rawValue, err := ite.Value()
			if err != nil {
				return fmt.Errorf("could not get raw SubsecTimeOriginal value")
			}

			val, ok := rawValue.(string)
			if !ok {
				return fmt.Errorf("OffsetTimeOriginal was not in expected format: %#v", rawValue)
			}

			val = strings.Replace(val, ":", "", 1)
			if len(val) != 5 {
				return fmt.Errorf("OffsetTimeOriginal was not of the expected length: %#v", rawValue)
			}

			offset, err = time.Parse("-0700", val)
			if err != nil {
				return fmt.Errorf("failed to parse OffsetTimeOriginal: %s", err)
			}
		}

		return nil
	}

	err = index.RootIfd.EnumerateTagsRecursively(cb)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to walk exif data tree: %s", err)
	}

	return time.Date(
		dateTimeOriginal.Year(),
		dateTimeOriginal.Month(),
		dateTimeOriginal.Day(),
		dateTimeOriginal.Hour(),
		dateTimeOriginal.Minute(),
		dateTimeOriginal.Second(),
		dateTimeOriginalSubSec,
		offset.Location(),
	).UTC(), nil
}
