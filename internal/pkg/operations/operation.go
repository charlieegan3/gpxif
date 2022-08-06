package operations

// Operation describes a set of related changes to an image.
type Operation struct {
	// Reason is why the operation will be run
	Reason string
	// IFDPath is the path to the IFD to operate on in the dataset
	IFDPath string
	// Fields is the desired state of some EXIF fields
	Fields map[string]interface{}
}
