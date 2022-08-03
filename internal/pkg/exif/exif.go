package exif

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	exif "github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	jpeg "github.com/dsoprea/go-jpeg-image-structure/v2"
)

func SetKeyString(image, key, value string) error {
	parser := jpeg.NewJpegMediaParser()
	intfc, err := parser.ParseFile(image)
	if err != nil {
		return fmt.Errorf("failed to parse %s as JPEG file: %v", image, err)
	}

	sl := intfc.(*jpeg.SegmentList)

	rootIb, err := sl.ConstructExifBuilder()
	if err != nil {
		im, err := exifcommon.NewIfdMappingWithStandard()
		if err != nil {
			return fmt.Errorf("failed to create new IFD mapping from scratch: %v", err)
		}
		ti := exif.NewTagIndex()
		if err := exif.LoadStandardTags(ti); err != nil {
			return fmt.Errorf("failed to load standard tags: %v", err)
		}

		rootIb = exif.NewIfdBuilder(
			im,
			ti,
			exifcommon.IfdStandardIfdIdentity,
			exifcommon.EncodeDefaultByteOrder,
		)
	}

	ifdPath := "IFD0"
	ifdIb, err := exif.GetOrCreateIbFromRootIb(rootIb, ifdPath)
	if err != nil {
		return fmt.Errorf("failed to get or create IB: %v", err)
	}

	if err := ifdIb.SetStandardWithName(key, value); err != nil {
		return fmt.Errorf("failed to set tag: %v", err)
	}

	if err := sl.SetExif(rootIb); err != nil {
		return fmt.Errorf("failed to set EXIF in rootIb: %v", err)
	}

	b := new(bytes.Buffer)
	if err := sl.Write(b); err != nil {
		return fmt.Errorf("failed to create JPEG data: %v", err)
	}

	if err := ioutil.WriteFile(image, b.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write JPEG file: %v", err)
	}

	return nil
}

func GetKeyString(image, key string) (string, error) {
	b, err := ioutil.ReadFile(image)
	if err != nil {
		return "", fmt.Errorf("failed to read image file: %w", err)
	}

	rawExif, err := exif.SearchAndExtractExif(b)
	if err == exif.ErrNoExif {
		return "", fmt.Errorf("no exif data found")
	} else if err != nil {
		return "", fmt.Errorf("failed to get raw exif data: %s", err)
	}

	im, err := exifcommon.NewIfdMappingWithStandard()
	if err != nil {
		return "", fmt.Errorf("failed to create idfmapping: %s", err)
	}

	ti := exif.NewTagIndex()

	_, index, err := exif.Collect(im, ti, rawExif)
	if err != nil {
		return "", fmt.Errorf("failed to collect exif data: %s", err)
	}

	var value string
	cb := func(ifd *exif.Ifd, ite *exif.IfdTagEntry) error {
		if ite.TagName() == key {
			rawValue, err := ite.Value()
			if err != nil {
				return fmt.Errorf("could not get raw value")
			}

			var ok bool
			value, ok = rawValue.(string)
			if !ok {
				return fmt.Errorf("value was not in expected format: %#v", rawValue)
			}
		}

		return nil
	}

	err = index.RootIfd.EnumerateTagsRecursively(cb)
	if err != nil {
		return "", fmt.Errorf("failed to walk exif data tree: %s", err)
	}

	return value, nil
}

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
		if ite.TagName() == "SubSecTimeOriginal" {
			rawValue, err := ite.Value()
			if err != nil {
				return fmt.Errorf("could not get raw SubSecTimeOriginal value")
			}

			val, ok := rawValue.(string)
			if !ok {
				return fmt.Errorf("SubSecTimeOriginal was not in expected format: %#v", rawValue)
			}

			intValue, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse SubSecTimeOriginal value: %w", err)
			}
			dateTimeOriginalSubSec = int(intValue)
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
