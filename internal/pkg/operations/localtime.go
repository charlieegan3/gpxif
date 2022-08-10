package operations

import (
	"fmt"
	"github.com/charlieegan3/gpxif/internal/pkg/exif"
	"github.com/charlieegan3/gpxif/internal/pkg/gpx"
	"github.com/zsefvlol/timezonemapper"
	"time"
)

func CheckLocalTime(imageFile string, g *gpx.GPXDataset) ([]Operation, error) {
	var operations []Operation

	// get the utc time for the image, if the image has an offset then this is used to calculate the utc time
	// if no offset is set, then the time is assumed to be UTC
	utcTime, err := exif.GetUTC(imageFile)
	if err != nil {
		return operations, fmt.Errorf("failed to get UTC time for image: %s", err)
	}

	// find the nearest point from the GPX track for that UTC time
	p, err := g.AtTime(utcTime)
	if err != nil {
		return operations, fmt.Errorf("failed to get point for time: %s\n\n", err)
	}

	// calculate the local time for the image from the UTC time and the GPS location
	location, err := time.LoadLocation(timezonemapper.LatLngToTimezoneString(p.Latitude, p.Longitude))
	if err != nil {
		return operations, fmt.Errorf("failed to parse location from GPS point: %s", err)
	}
	local := utcTime
	local = local.In(location)

	// check that the DateTimeOriginal and Offset are set to show local time
	expectedDateTime := local.Format("2006:01:02 15:04:05")
	expectedOffset := local.Format("-07:00")
	expectedSubSec := fmt.Sprintf("%d", local.Nanosecond()/1000000)

	currentDateTime, err := exif.GetKey(imageFile, "IFD/Exif", "DateTimeOriginal")
	if err != nil {
		return operations, fmt.Errorf("failed to get DateTimeOriginal: %v", err)
	}
	currentSubSecTime, err := exif.GetKey(imageFile, "IFD/Exif", "SubSecTimeOriginal")
	if err != nil {
		currentSubSecTime = ""
	}
	currentOffset, err := exif.GetKey(imageFile, "IFD/Exif", "OffsetTimeOriginal")
	if err != nil {
		currentOffset = ""
	}

	trigger := false
	o := Operation{
		Reason:  "DateTimeOriginal data was not in local time",
		IFDPath: "IFD/Exif",
		Fields:  map[string]interface{}{},
	}
	if currentDateTime != expectedDateTime {
		trigger = true
		o.Fields["DateTimeOriginal"] = expectedDateTime
	}
	if currentOffset != expectedOffset {
		trigger = true
		o.Fields["OffsetTimeOriginal"] = expectedOffset
	}
	if currentSubSecTime != expectedSubSec {
		trigger = true
		o.Fields["SubSecTimeOriginal"] = expectedSubSec
	}

	if trigger {
		operations = append(operations, o)
	}

	return operations, nil
}
