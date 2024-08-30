package gpxfetch

import (
	"bytes"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"sort"
	"text/template"
	"time"

	"github.com/charlieegan3/gpxif/internal/pkg/config"
	"github.com/charlieegan3/gpxif/internal/pkg/exif"
	"github.com/charlieegan3/gpxif/internal/pkg/gpx"
	"github.com/charlieegan3/gpxif/internal/pkg/utils"
)

func ForImages(cfg config.Config, sourceDir string) (gpx.GPXDataset, error) {
	var gpxDataset gpx.GPXDataset

	rangeStart, rangeEnd, err := determineTimeRange(sourceDir)
	if err != nil {
		return gpxDataset, fmt.Errorf("failed to determine time range for images: %w", err)
	}

	rangeStartDate := rangeStart.Format("2006-01-02")
	rangeEndDate := rangeEnd.Format("2006-01-02")

	var buf bytes.Buffer

	t, err := template.New("url").Parse(cfg.GPXSource.URLTemplate)
	if err != nil {
		return gpxDataset, fmt.Errorf("failed to parse URL template: %w", err)
	}
	err = t.Execute(&buf, struct{ From, To string }{From: rangeStartDate, To: rangeEndDate})
	if err != nil {
		return gpxDataset, fmt.Errorf("failed to populate URL template: %w", err)
	}

	req, err := http.NewRequest("GET", buf.String(), nil)
	if err != nil {
		return gpxDataset, fmt.Errorf("failed to create request to source GPX data: %w", err)
	}
	req.SetBasicAuth(cfg.GPXSource.Username, cfg.GPXSource.Password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return gpxDataset, fmt.Errorf("failed to get GPX data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return gpxDataset, fmt.Errorf("failed to get GPX data, status code %d", resp.StatusCode)
	}

	gpxDataset, err = gpx.NewGPXDatasetFromReader(resp.Body)
	if err != nil {
		return gpxDataset, fmt.Errorf("failed to parse GPX data: %w", err)
	}

	return gpxDataset, nil
}

// determineTimeRange returns the earliest and latest times from a set of images
func determineTimeRange(sourceDir string) (time.Time, time.Time, error) {
	var err error
	var start, end time.Time

	var files []fs.FileInfo
	files, err = ioutil.ReadDir(sourceDir)
	if err != nil {
		return start, end, fmt.Errorf("failed to list files in source directory: %s", err)
	}

	var utcTimes []time.Time
	for _, f := range files {
		if !utils.IsJPEGFile(f.Name()) {
			continue
		}

		utcTime, err := exif.GetUTC(sourceDir + "/" + f.Name())
		if err != nil {
			return start, end, fmt.Errorf("failed to determine UTC time for %s: %w", f.Name(), err)
		}

		utcTimes = append(utcTimes, utcTime)
	}

	if len(utcTimes) < 1 {
		return start, end, fmt.Errorf("no images with UTC times found in source directory")
	}

	sort.Slice(utcTimes, func(i, j int) bool {
		return utcTimes[i].Before(utcTimes[j])
	})

	return utcTimes[0], utcTimes[len(utcTimes)-1], nil
}
