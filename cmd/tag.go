package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"strings"

	"github.com/charlieegan3/gpxif/internal/pkg/gpx"
	"github.com/charlieegan3/gpxif/internal/pkg/operations"
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
			if !strings.HasSuffix(strings.ToLower(f.Name()), ".jpg") {
				fmt.Println(f.Name(), "skipped")
				continue
			}

			var ops []operations.Operation

			gpsOperations, err := operations.CheckGPSData(imageSource+"/"+f.Name(), g)
			if err != nil {
				log.Fatalf("failed to determine GPS operations for %s: %s", f.Name(), err)
			}
			ops = append(ops, gpsOperations...)

			timeOperations, err := operations.CheckLocalTime(imageSource+"/"+f.Name(), g)
			if err != nil {
				log.Fatalf("failed to determine local time operations for %s: %s", f.Name(), err)
			}

			ops = append(ops, timeOperations...)

			if len(ops) == 0 {
				continue
			}

			fmt.Println("Updates to ", f.Name())

			for _, op := range ops {
				fmt.Printf("  %s\n", op.Reason)
				for k, v := range op.Fields {
					fmt.Printf("    Set %q to %v\n", k, v)
				}

				if !dryRun {
					err := op.Execute(imageSource + "/" + f.Name())
					if err != nil {
						log.Fatalf("failed operation: %s", err)
					}
				}
			}
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
	err := tagCmd.MarkFlagRequired("images")
	if err != nil {
		log.Fatalf("Failed to mark images flag required: %s", err)
	}
	err = tagCmd.MarkFlagRequired("gpx")
	if err != nil {
		log.Fatalf("Failed to mark gpx flag required: %s", err)
	}
}
