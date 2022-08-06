package operations

import (
	"fmt"
	dectofrac "github.com/av-elier/go-decimal-to-rational"
	"github.com/charlieegan3/gpxif/internal/pkg/exif"
	"github.com/charlieegan3/gpxif/internal/pkg/gpx"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	"strings"
)

func CheckGPSData(imageFile string, g gpx.GPXDataset) ([]Operation, error) {
	var operations []Operation

	gpsLatitude, err := exif.GetKey(imageFile, "IFD/GPSInfo", "GPSLatitude")
	if err != nil && !strings.Contains(err.Error(), "tag not found") {
		return operations, fmt.Errorf("failed to get GPSLatitude: %w", err)
	}

	// assume that the other GPS values are set
	if gpsLatitude != nil {
		return operations, nil
	}

	// get the UTC time of the image
	utcTime, err := exif.GetUTC(imageFile)
	if err != nil {
		return operations, fmt.Errorf("failed to determine UTC time for image: %w", err)
	}

	// find the point in the gpx dataset that matches the UTC time of the image
	point, err := g.AtTime(utcTime)
	if err != nil {
		return operations, fmt.Errorf("failed to find point at image UTC time: %w", err)
	}

	// get the values from the point in the correct format to set in EXIF
	gpsLatitudeRational := exif.RationalDegreesMinutesSecondsFromDecimal(point.Latitude)
	gpsLongitudeRational := exif.RationalDegreesMinutesSecondsFromDecimal(point.Longitude)

	var gpsLatitudeRef, gpsLongitudeRef string
	if point.Latitude > 0 {
		gpsLatitudeRef = "N"
	} else {
		gpsLatitudeRef = "S"
	}
	if point.Longitude > 0 {
		gpsLongitudeRef = "E"
	} else {
		gpsLongitudeRef = "W"
	}

	altitude := dectofrac.NewRatP(point.Elevation.Value(), 0.0001)
	altitudeRational := []exifcommon.Rational{
		{
			Numerator:   uint32(altitude.Num().Int64()),
			Denominator: uint32(altitude.Denom().Int64()),
		},
	}

	// set the values in the EXIF
	operations = append(operations, Operation{
		Reason:  "GPS data not found in EXIF",
		IFDPath: "IFD/GPSInfo",
		Fields: map[string]interface{}{
			"GPSLatitude":     gpsLatitudeRational,
			"GPSLatitudeRef":  gpsLatitudeRef,
			"GPSLongitude":    gpsLongitudeRational,
			"GPSLongitudeRef": gpsLongitudeRef,
			"GPSAltitude":     altitudeRational,
		},
	})

	return operations, nil
}
