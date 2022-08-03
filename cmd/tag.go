/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/charlieegan3/gpxif/internal/pkg/exif"
	"github.com/charlieegan3/gpxif/internal/pkg/gpx"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"strings"
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

			utcTime, err := exif.GetUTC(imageSource + "/" + f.Name())
			if err != nil {
				log.Fatalf("Failed to get UTC time for image: %s", err)
			}

			fmt.Println(utcTime)
			p, err := g.AtTime(utcTime)
			if err != nil {
				log.Fatalf("Failed to get point for time: %s", err)
			}
			fmt.Println(p.Latitude, p.Longitude)
		}
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
	tagCmd.MarkFlagRequired("gpx")
	if err != nil {
		log.Fatalf("Failed to mark gpx flag required: %s", err)
	}
}
