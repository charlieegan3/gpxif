package exif

import (
	"encoding/binary"
	"fmt"
	jpegstructure "github.com/dsoprea/go-jpeg-image-structure/v2"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	exif "github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
)

// getIndexedTagFromName looks up tag index values to use for supplied tags. When we have a new tag that's not in the
// current file, then we need to look up where it should go in the EXIF tree
func getIndexedTagFromName(k string) (*exifcommon.IfdIdentity, *exif.IndexedTag, error) {
	tag_paths := []*exifcommon.IfdIdentity{
		exifcommon.IfdStandardIfdIdentity,
		exifcommon.IfdExifStandardIfdIdentity,
		exifcommon.IfdExifIopStandardIfdIdentity,
		exifcommon.IfdGpsInfoStandardIfdIdentity,
		exifcommon.Ifd1StandardIfdIdentity,
	}
	ti := exif.NewTagIndex()

	for _, id := range tag_paths {
		t, err := ti.GetWithName(id, k)

		if err != nil {
			continue
		}

		return id, t, nil
	}

	return nil, nil, fmt.Errorf("unrecognized tag, %s", k)
}

func SetKeyString(image, key, value string) error {
	jmp := jpegstructure.NewJpegMediaParser()

	intfc, err := jmp.ParseFile(image)
	if err != nil {
		return fmt.Errorf("failed to parse image: %s", err)
	}

	sl := intfc.(*jpegstructure.SegmentList)

	rootIb, err := sl.ConstructExifBuilder()
	if err != nil {
		return fmt.Errorf("failed to construct exif builder: %s", err)
	}

	ifdPath := "IFD/Exif"

	childIb, err := exif.GetOrCreateIbFromRootIb(rootIb, ifdPath)
	if err != nil {
		return fmt.Errorf("failed to get child ifd builder: %s", err)
	}

	_, it, err := getIndexedTagFromName(key)
	if err != nil {
		return fmt.Errorf("failed to lookup indexed tag from name: %s", err)
	}

	childIb.Set(exif.NewBuilderTag(
		ifdPath,
		it.Id,
		// TODO we can only set ascii values
		exifcommon.TypeAscii,
		exif.NewIfdBuilderTagValueFromBytes([]byte(value)),
		binary.BigEndian,
	))

	fmt.Println(image, key, value)
	childIb.PrintTagTree()

	err = sl.SetExif(rootIb)
	if err != nil {
		return fmt.Errorf("failed to set exif data: %s", err)
	}

	f, err := os.Create(image)
	if err != nil {
		return fmt.Errorf("failed to get file handle for image: %s", err)
	}
	defer f.Close()

	err = sl.Write(f)

	return err
}

func SetLocalTime(image string, localTime time.Time) error {
	dateTime := localTime.Format("2006-01-02 15:04:05")
	offset := localTime.Format("-07:00")
	subSec := localTime.Format("200")

	err := SetKeyString(image, "DateTimeOriginal", dateTime)
	if err != nil {
		return fmt.Errorf("failed to set DateTimeOriginal: %v", err)
	}
	err = SetKeyString(image, "SubSecTimeOriginal", subSec)
	if err != nil {
		return fmt.Errorf("failed to set SubSecTimeOriginal: %v", err)
	}
	err = SetKeyString(image, "OffsetTimeOriginal", offset)
	if err != nil {
		return fmt.Errorf("failed to set OffsetTimeOriginal: %v", err)
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
		dateTimeOriginalSubSec*1000000,
		offset.Location(),
	).UTC(), nil
}
