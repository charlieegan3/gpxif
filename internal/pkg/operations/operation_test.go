package operations

import (
	"github.com/charlieegan3/gpxif/internal/pkg/exif"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestOperationExecute(t *testing.T) {
	testCases := map[string]struct {
		Image      string
		Operations []Operation
	}{
		"example update to time data": {
			Image: "../exif/fixtures/x100f.jpg",
			Operations: []Operation{
				{
					Reason:  "DateTimeOriginal data was not in local time",
					IFDPath: "IFD/Exif",
					Fields: map[string]interface{}{
						"DateTimeOriginal":   "2022:08:03 18:57:55",
						"OffsetTimeOriginal": "+01:00",
						"SubSecTimeOriginal": "0",
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			imageCopy, err := ioutil.TempFile(".", "image_")
			require.NoError(t, err)
			defer os.Remove(imageCopy.Name())

			imageFile, err := os.Open(testCase.Image)
			require.NoError(t, err)

			_, err = io.Copy(imageCopy, imageFile)
			require.NoError(t, err)

			for _, op := range testCase.Operations {
				err := op.Execute(imageCopy.Name())
				require.NoError(t, err)

				for k, v := range op.Fields {
					value, err := exif.GetKey(imageCopy.Name(), op.IFDPath, k)
					require.NoError(t, err)
					assert.Equal(t, v, value)
				}
			}
		})
	}
}
