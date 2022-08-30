package operations

import (
	"fmt"
	"github.com/charlieegan3/gpxif/internal/pkg/exif"
	"github.com/djherbis/times"
)

func CheckModTime(imageFile string) ([]Operation, error) {
	var operations []Operation

	utcTime, err := exif.GetUTC(imageFile)
	if err != nil {
		return operations, fmt.Errorf("failed to get UTC time for image: %s", err)
	}

	t, err := times.Stat(imageFile)
	if err != nil {
		return operations, fmt.Errorf("failed to get mtime for image: %s", err)
	}

	format := "2006 02 01 15 04"
	if t.ModTime().UTC().Format(format) != utcTime.Format(format) {
		operations = append(operations, Operation{
			Reason:  "mtime != utc time",
			ModTime: true,
		})
	}

	return operations, nil
}
