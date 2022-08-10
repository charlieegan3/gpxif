package gpx

import (
	"fmt"
	"github.com/tkrajina/gpxgo/gpx"
	"io"
	"sort"
	"time"
)

type GPXDataset struct {
	data []*gpx.GPX
}

func (g *GPXDataset) InRange(t time.Time) bool {
	if len(g.data) < 1 {
		return false
	}

	var allPoints []gpx.GPXPoint
	for _, d := range g.data {
		for _, track := range d.Tracks {
			for _, segment := range track.Segments {
				for _, point := range segment.Points {
					allPoints = append(allPoints, point)
				}
			}
		}
	}

	if len(allPoints) < 1 {
		return false
	}

	sort.Slice(allPoints, func(i, j int) bool { return allPoints[i].Timestamp.Before(allPoints[j].Timestamp) })

	if t.Before(allPoints[0].Timestamp) || t.After(allPoints[len(allPoints)-1].Timestamp) {
		return false
	}

	return true
}

func (g *GPXDataset) AtTime(t time.Time) (gpx.GPXPoint, error) {
	if !g.InRange(t) {
		return gpx.GPXPoint{}, fmt.Errorf("time %s was out of range of loaded files", t)
	}

	var closestPoint gpx.GPXPoint

	minDiff := time.Hour * 24 * 365
	match := false

	for _, d := range g.data {
		for _, track := range d.Tracks {
			for _, segment := range track.Segments {
				for _, point := range segment.Points {

					var diff time.Duration

					if point.Timestamp.Before(t) {
						diff = t.Sub(point.Timestamp)
					} else {
						diff = point.Timestamp.Sub(t)
					}

					if diff < minDiff {
						match = true
						minDiff = diff
						closestPoint = point
					}
				}
			}
		}
	}

	if !match {
		return gpx.GPXPoint{}, fmt.Errorf("no match found for time %s", t)
	}

	return closestPoint, nil
}

func NewGPXDatasetFromFile(files ...string) (GPXDataset, error) {
	ds := GPXDataset{}

	for _, f := range files {
		data, err := gpx.ParseFile(f)
		if err != nil {
			return GPXDataset{}, fmt.Errorf("failed to parse gpx data from file %s: %w", f, err)
		}

		ds.data = append(ds.data, data)
	}

	return ds, nil
}

func NewGPXDatasetFromReader(reader io.Reader) (GPXDataset, error) {
	ds := GPXDataset{}

	b, err := io.ReadAll(reader)
	if err != nil {
		return GPXDataset{}, fmt.Errorf("failed to data from reader: %w", err)
	}

	rawData, err := gpx.ParseBytes(b)
	if err != nil {
		return GPXDataset{}, fmt.Errorf("failed to parse gpx data from file: %w", err)
	}

	ds.data = append(ds.data, rawData)

	return ds, nil
}
