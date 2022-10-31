package cmd

import (
	"fmt"
	"github.com/charlieegan3/gpxif/internal/pkg/config"
	"github.com/charlieegan3/gpxif/internal/pkg/gpxfetch"
	"github.com/mitchellh/go-homedir"
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

		autoSource, err := cmd.Flags().GetBool("auto")
		if err != nil {
			log.Fatalf("Failed to get auto flag: %s", err)
		}

		var g *gpx.GPXDataset

		if autoSource {
			path, err := homedir.Expand("~/.gpxif")
			if err != nil {
				log.Fatalf("error expanding homedir: %v", err)
			}

			cfg, err := config.Load(path)
			if err != nil {
				log.Fatalf("failed to load config: %s", err)
			}

			fmt.Println("Auto sourcing GPX data")
			fmt.Println("GPX Source:", cfg.GPXSource.URLTemplate)
			fmt.Println("GPX Source Username:", cfg.GPXSource.Username)

			autoDs, err := gpxfetch.ForImages(cfg, imageSource)
			if err != nil {
				log.Fatalf("failed to auto source gpx data: %s", err)
			}
			g = &autoDs
		} else {
			gpxSource, err := cmd.Flags().GetString("gpx")
			if err != nil {
				log.Fatalf("Failed to get gpxSource flag: %s", err)
			}

			fileDs, err := gpx.NewGPXDatasetFromDisk(gpxSource)
			if err != nil {
				log.Fatalf("Failed to create GPX dataset: %s", err)
			}
			g = &fileDs
		}

		fmt.Println("Dry Run: ", dryRun)
		fmt.Println("Image Source: ", imageSource)
		fmt.Println("---")

		files, err := ioutil.ReadDir(imageSource)
		if err != nil {
			log.Fatalf("Failed to list files in images directory: %s", err)
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

			// TODO: these need to be last since they depend on data set in other operations
			modTimeOperations, err := operations.CheckModTime(imageSource + "/" + f.Name())
			if err != nil {
				log.Fatalf("failed to determine mtime operations for %s: %s", f.Name(), err)
			}
			ops = append(ops, modTimeOperations...)

			if len(ops) == 0 {
				continue
			}

			fmt.Println("Updates to", f.Name())

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
	tagCmd.Flags().Bool(
		"auto",
		false,
		"Automatically determine the GPX data based on image timestamps",
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
}
