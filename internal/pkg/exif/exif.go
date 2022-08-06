package exif

import (
	"fmt"
	jpegstructure "github.com/dsoprea/go-jpeg-image-structure/v2"
	"io/ioutil"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	exif "github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
)

// GetKey extracts an abitrary key from the image's EXIF data
func GetKey(image, targetIFDPath, key string) (interface{}, error) {
	jmp := jpegstructure.NewJpegMediaParser()

	intfc, err := jmp.ParseFile(image)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image: %s", err)
	}

	rootIfd, _, err := intfc.Exif()
	if err != nil {
		return nil, fmt.Errorf("failed to get root ifd: %s", err)
	}

	_, it, err := getIndexedTagFromName(key)
	if err != nil {
		return "", fmt.Errorf("failed to lookup indexed tag from name: %w", err)
	}

	var retValue interface{}
	for _, c := range rootIfd.Children() {
		if c.IfdIdentity().String() == targetIFDPath {
			results, err := c.FindTagWithId(it.Id)
			if err != nil {
				return "", fmt.Errorf("failed to find tag %s: %w", key, err)
			}

			if len(results) > 0 {
				retValue, err = results[0].Value()
				if err != nil {
					return "", fmt.Errorf("failed to get value for key %s: %w", key, err)
				}
			}
		}
	}

	return retValue, nil
}

// SetKey sets a key value of Ascii or []Rational in the exif data at the specified path
func SetKey(image, targetIFDPath, key string, value any) error {
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

	rootIfd, _, err := intfc.Exif()
	if err != nil {
		return fmt.Errorf("failed to get root ifd: %s", err)
	}

	// strip the thumbnail since it gets messed with by go-exif
	childIb, err := exif.GetOrCreateIbFromRootIb(rootIb, "IFD1")
	if err != nil {
		return fmt.Errorf("failed to get child ifd builder: %s", err)
	}
	childIb.DeleteAll(0x0201)
	childIb.DeleteAll(0x0202)

	// yeet these invalid values
	childIb, err = exif.GetOrCreateIbFromRootIb(rootIb, "IFD/Exif")
	if err != nil {
		return fmt.Errorf("failed to get child ifd builder for IFD/Exif: %s", err)
	}
	childIb.DeleteAll(0xa301)
	childIb.DeleteAll(0xa300)

	// set the value we want to set
	childIb, err = exif.GetOrCreateIbFromRootIb(rootIb, targetIFDPath)
	if err != nil {
		return fmt.Errorf("failed to get child ifd builder: %s", err)
	}

	_, it, err := getIndexedTagFromName(key)
	if err != nil {
		return fmt.Errorf("failed to lookup indexed tag from name: %s", err)
	}

	enc := exifcommon.NewValueEncoder(rootIfd.ByteOrder())
	data, err := enc.Encode(value)
	if err != nil {
		return fmt.Errorf("failed to encode value: %s", err)
	}

	valueType := exifcommon.TypeAscii
	switch value.(type) {
	case string:
		valueType = exifcommon.TypeAscii
	case []exifcommon.Rational:
		valueType = exifcommon.TypeRational
	default:
		return fmt.Errorf("unsupported value type: %s", reflect.TypeOf(value))
	}

	err = childIb.Set(exif.NewBuilderTag(
		targetIFDPath,
		it.Id,
		valueType,
		exif.NewIfdBuilderTagValueFromBytes(data.Encoded),
		rootIfd.ByteOrder(),
	))
	if err != nil {
		return fmt.Errorf("failed to set value: %s", err)
	}

	// write the data back to the file
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

	// TODO I think this value is meant to be in milliseconds
	subSec := fmt.Sprintf("%d", localTime.Nanosecond()/1000000)

	err := SetKey(image, "IFD/Exif", "DateTimeOriginal", dateTime)
	if err != nil {
		return fmt.Errorf("failed to set DateTimeOriginal: %v", err)
	}
	err = SetKey(image, "IFD/Exif", "SubSecTimeOriginal", subSec)
	if err != nil {
		return fmt.Errorf("failed to set SubSecTimeOriginal: %v", err)
	}
	err = SetKey(image, "IFD/Exif", "OffsetTimeOriginal", offset)
	if err != nil {
		return fmt.Errorf("failed to set OffsetTimeOriginal: %v", err)
	}

	return nil
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

func RationalDegreesMinutesSecondsFromDecimal(decimal float64) []exifcommon.Rational {
	decimal = math.Abs(decimal)

	minutes := math.Abs((decimal - math.Floor(decimal)) * 60)
	seconds := (minutes - math.Floor(minutes)) * 60

	return []exifcommon.Rational{
		{
			Numerator:   uint32(math.Abs(math.Floor(decimal))),
			Denominator: 1,
		},
		{
			Numerator:   uint32(math.Floor(minutes)),
			Denominator: 1,
		},
		{
			Numerator:   uint32(math.Floor(seconds * 100)),
			Denominator: 100,
		},
	}
}

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
