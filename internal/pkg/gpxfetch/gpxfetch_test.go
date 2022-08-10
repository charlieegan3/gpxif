package gpxfetch

import (
	"fmt"
	"github.com/charlieegan3/gpxif/internal/pkg/config"
	"github.com/charlieegan3/gpxif/internal/pkg/gpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestForImages(t *testing.T) {
	rawGPXData := `
<?xml version="1.0" encoding="UTF-8"?>
<gpx xmlns="http://www.topografix.com/GPX/1/0" version="1.0">
	<trk>
		<name>2022-08-03 to 2022-08-03</name>
		<trkseg>
			<trkpt lat="51.56734" lon="-0.13843">
				<ele>75</ele>
				<time>2022-08-03T01:18:07Z</time>
			</trkpt>
			<trkpt lat="51.56734" lon="-0.13843">
				<ele>75</ele>
				<time>2022-08-03T23:18:07Z</time>
			</trkpt>
		</trkseg>
	</trk>
</gpx>
`
	expectedDs, err := gpx.NewGPXDatasetFromReader(strings.NewReader(rawGPXData))
	require.NoError(t, err)

	sourceDir := "../exif/fixtures"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok {
			t.Fatalf("missing basic auth settings")
		}
		if user != "user" && pass != "pass" {
			t.Fatalf("incorrect basic auth settings")
		}

		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")
		if from != "2022-01-21" || to != "2022-08-03" {
			t.Fatalf("unexpected query params: from=%s, to=%s", from, to)
		}

		w.Write([]byte(rawGPXData))
	}))
	defer ts.Close()

	gpxDs, err := ForImages(
		config.Config{
			GPXSource: config.GPXSource{
				URLTemplate: fmt.Sprintf("%s?from={{ .From }}&to={{ .To }}", ts.URL),
				Username:    "user",
				Password:    "pass",
			},
		},
		sourceDir,
	)
	require.NoError(t, err)

	assert.Equal(t, expectedDs, gpxDs)
}

func TestDetermineRange(t *testing.T) {
	testCases := map[string]struct {
		SourceDir     string
		ExpectedStart time.Time
		ExpectedEnd   time.Time
	}{
		"check exif fixture dir": {
			SourceDir:     "../exif/fixtures",
			ExpectedStart: time.Date(2022, time.January, 21, 9, 9, 0, 97000000, time.UTC),
			ExpectedEnd:   time.Date(2022, time.August, 3, 17, 57, 55, 0, time.UTC),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			start, end, err := determineTimeRange(testCase.SourceDir)
			require.NoError(t, err)

			assert.Equal(t, testCase.ExpectedEnd, end)
			assert.Equal(t, testCase.ExpectedStart, start)
		})
	}
}
