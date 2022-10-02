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

func (g *GPXDataset) AllPoints() []gpx.GPXPoint {
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

	sort.Slice(allPoints, func(i, j int) bool { return allPoints[i].Timestamp.Before(allPoints[j].Timestamp) })

	return allPoints
}

// InRange returns true if the given time is within the range of the loaded GPX data.
// The second return value is whether the time is before or after the range
func (g *GPXDataset) InRange(t time.Time) (bool, bool) {
	if len(g.data) < 1 {
		return false, false
	}

	allPoints := g.AllPoints()

	if len(allPoints) < 1 {
		return false, false
	}

	if t.Before(allPoints[0].Timestamp) {
		return false, true
	}
	if t.After(allPoints[len(allPoints)-1].Timestamp) {
		return false, false
	}

	return true, false
}

// AtTime returns the closest point for a given time.
// If the time is not in the range of the points, then the first or last point is used.
// However, the offset of such matches are limited to 24hrs.
func (g *GPXDataset) AtTime(t time.Time) (gpx.GPXPoint, error) {
	inRange, before := g.InRange(t)
	if !inRange {
		allPoints := g.AllPoints()
		if len(allPoints) == 0 {
			return gpx.GPXPoint{}, fmt.Errorf("no points in dataset")
		}
		if before {
			candidatePoint := allPoints[0]
			if candidatePoint.Timestamp.Sub(t) < 24*time.Hour {
				return candidatePoint, nil
			}

			return gpx.GPXPoint{}, fmt.Errorf("out of range of loaded files: %v", t)
		} else {
			candidatePoint := allPoints[len(allPoints)-1]
			if t.Sub(candidatePoint.Timestamp) < 24*time.Hour {
				return candidatePoint, nil
			}
			return gpx.GPXPoint{}, fmt.Errorf("out of range of loaded files: %v", t)
		}
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
