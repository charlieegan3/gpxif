package utils

import "testing"

func TestIsJPEGFile(t *testing.T) {
	testCases := map[string]struct {
		filename string
		isJPEG   bool
	}{
		"jpeg": {
			filename: "foo.jpeg",
			isJPEG:   true,
		},
		"jpg": {
			filename: "foo.jpg",
			isJPEG:   true,
		},
		"JPG": {
			filename: "foo.JPG",
			isJPEG:   true,
		},
		"JPEG": {
			filename: "foo.JPEG",
			isJPEG:   true,
		},
		"png": {
			filename: "foo.png",
			isJPEG:   false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if got := IsJPEGFile(tc.filename); got != tc.isJPEG {
				t.Errorf("want %v, got %v", tc.isJPEG, got)
			}
		})
	}
}
