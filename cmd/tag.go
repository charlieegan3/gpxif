package cmd

import (
	"fmt"
	"github.com/charlieegan3/gpxif/internal/pkg/exif"
	"github.com/charlieegan3/gpxif/internal/pkg/gpx"
	"github.com/spf13/cobra"
	"github.com/zsefvlol/timezonemapper"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

// tagCmd represents the tag command
var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "tag takes GPX data and images and adds EXIF data to the images using UTC timestamps as a cross reference",
	Run: func(cmd *cobra.Command, args []string) {
		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			log.Fatalf("Failed to get dry-run flag: %s", err)
		}

		imageSource, err := cmd.Flags().GetString("images")
		if err != nil {
			log.Fatalf("Failed to get imageSource flag: %s", err)
		}
		imageSource = strings.TrimSuffix(imageSource, "/")

		gpxSource, err := cmd.Flags().GetString("gpx")
		if err != nil {
			log.Fatalf("Failed to get gpxSource flag: %s", err)
		}

		fmt.Println("Dry Run: ", dryRun)
		fmt.Println("Image Source: ", imageSource)
		fmt.Println("GPX Source: ", gpxSource)
		fmt.Println("---")

		files, err := ioutil.ReadDir(imageSource)
		if err != nil {
			log.Fatalf("Failed to list files in images directory: %s", err)
		}

		g, err := gpx.NewGPXDatasetFromFile(gpxSource)
		if err != nil {
			log.Fatalf("Failed to create GPX dataset: %s", err)
		}

		for _, f := range files {
			fmt.Println("Processing", f.Name())

			imageFileName := imageSource + "/" + f.Name()

			// get the utc time for the image, if the image has an offset then this is used to calculate the utc time
			// if no offset is set, then the time is assumed to be UTC
			utcTime, err := exif.GetUTC(imageFileName)
			if err != nil {
				log.Fatalf("Failed to get UTC time for image: %s", err)
			}
			fmt.Println(utcTime)

			// find the nearest point from the GPX track for that UTC time
			p, err := g.AtTime(utcTime)
			if err != nil {
				log.Printf("Failed to get point for time: %s\n\n", err)
				continue
			}
			fmt.Println(p.Latitude, p.Longitude)

			// calculate the local time for the image from the UTC time and the GPS location
			location, err := time.LoadLocation(timezonemapper.LatLngToTimezoneString(p.Latitude, p.Longitude))
			if err != nil {
				log.Fatalf("Failed to parse location from GPS point: %s", err)
			}
			local := utcTime
			local = local.In(location)

			// check that the DateTimeOriginal and Offset are set to show local time
			expectedDateTime := local.Format("2006-01-02 15:04:05")
			expectedOffset := local.Format("-07:00")
			expectedSubSec := fmt.Sprintf("%d", local.Nanosecond()/1000000)

			currentDateTime, err := exif.GetKeyASCII(imageFileName, "DateTimeOriginal")
			if err != nil {
				log.Fatalf("failed to get DateTimeOriginal: %v", err)
			}
			currentSubSecTime, err := exif.GetKeyASCII(imageFileName, "SubSecTimeOriginal")
			if err != nil {
				log.Fatalf("failed to get SubSecTimeOriginal: %v", err)
			}
			currentOffset, err := exif.GetKeyASCII(imageFileName, "OffsetTimeOriginal")
			if err != nil {
				log.Fatalf("failed to get OffsetTimeOriginal: %v", err)
			}

			fmt.Println("DateTimeOriginal:", currentDateTime, "->", expectedDateTime)
			fmt.Println("SubSecTimeOriginal:", currentSubSecTime, "->", expectedSubSec)
			fmt.Println("OffsetTimeOriginal:", currentOffset, "->", expectedOffset)

			fmt.Println("")
		}

		// TODO: function to calc abs diff in duration between two times to handle diff limit
	},
}

func init() {
	rootCmd.AddCommand(tagCmd)

	tagCmd.Flags().Bool(
		"dry-run",
		false,
		"Don't update images, just print what would be done",
	)
	tagCmd.Flags().StringP(
		"images",
		"i",
		"",
		"Directory containing images to tag",
	)
	tagCmd.Flags().StringP(
		"gpx",
		"g",
		"",
		"GPX file containing timestamps",
	)
	tagCmd.Flags().StringP(
		"diff",
		"d",
		"5m",
		"Maximum permitted diff between GPS point and image UTC times",
	)
	err := tagCmd.MarkFlagRequired("images")
	if err != nil {
		log.Fatalf("Failed to mark images flag required: %s", err)
	}
	err = tagCmd.MarkFlagRequired("gpx")
	if err != nil {
		log.Fatalf("Failed to mark gpx flag required: %s", err)
	}
}
