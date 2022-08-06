# gpxif

CLI to update image EXIF locations and times using a GPX track.

Example usage:

```shell
go run main.go tag -i ~/Downloads/photos/ -g ~/Downloads/2022-08-01-to-2022-08-07.gpx 
```

Options:

* `-i` sets the source of images
* `-g` sets the source of the GPX file
* `--dry-run` will cause update operations to be printed without edits being made
