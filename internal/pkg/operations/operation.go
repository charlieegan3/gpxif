package operations

import (
	"fmt"
	"github.com/charlieegan3/gpxif/internal/pkg/exif"
)

// Operation describes a set of related changes to an image.
type Operation struct {
	// Reason is why the operation will be run
	Reason string
	// IFDPath is the path to the IFD to operate on in the dataset
	IFDPath string
	// Fields is the desired state of some EXIF fields
	Fields map[string]interface{}
}

func (o *Operation) Execute(image string) error {
	for k, v := range o.Fields {
		err := exif.SetKey(image, o.IFDPath, k, v)
		if err != nil {
			return fmt.Errorf("failed to set %v to %v: %s", k, v, err)
		}
	}

	return nil
}
