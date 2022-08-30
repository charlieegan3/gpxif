package operations

import (
	"fmt"
	"github.com/charlieegan3/gpxif/internal/pkg/exif"
	"os"
)

// Operation describes a set of related changes to an image.
type Operation struct {
	// Reason is why the operation will be run
	Reason string
	// IFDPath is the path to the IFD to operate on in the dataset
	IFDPath string
	// Fields is the desired state of some EXIF fields
	Fields map[string]interface{}

	// ModTime if set will trigger the operation exec to update the mtime of the
	// file to the DateTimeOriginal of the image.
	ModTime bool
}

func (o *Operation) Execute(image string) error {
	if o.ModTime {
		utcTime, err := exif.GetUTC(image)
		if err != nil {
			return fmt.Errorf("failed to get utc time: %w", err)
		}

		err = os.Chtimes(image, utcTime, utcTime)
		if err != nil {
			return fmt.Errorf("failed to set image mtime from utc DateTimeOriginal value: %w", err)
		}
		return nil
	}

	// otherwise, update the exif data
	for k, v := range o.Fields {
		err := exif.SetKey(image, o.IFDPath, k, v)
		if err != nil {
			return fmt.Errorf("failed to set %v to %v: %s", k, v, err)
		}
	}

	return nil
}
