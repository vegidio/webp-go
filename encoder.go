package webp

import (
	"fmt"
	"image"
	"image/draw"
	"io"
)

// Options represent the configuration options for encoding a HEIC image.
//
//   - Quality: Specifies the quality of the image, from 0-100, where 100 means lossless encoding. Higher values result
//     in better quality but bigger images (default 60).
type Options struct {
	Quality int
}

// Encode encodes an image into the HEIC format and writes it to the provided writer.
//
// Parameters:
//   - writer: The destination where the encoded HEIC image will be written.
//   - img: The input image to be encoded.
//   - options: A pointer to an Options struct that specifies encoding parameters. If nil, default values are used.
//
// Returns:
//   - An error if encoding or writing fails, otherwise nil.
func Encode(writer io.Writer, img image.Image, options *Options) error {
	// Convert the image to RGBA
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, rgba.Bounds(), img, bounds.Min, draw.Src)

	// Set default values for options if they are not set
	if options == nil {
		options = &Options{Quality: 60}
	}

	if options.Quality < 0 || options.Quality > 100 {
		return fmt.Errorf("quality must be between 0 and 100")
	}

	data, err := encodeHEIF(*rgba, *options)
	if err != nil {
		return err
	}

	if _, err = writer.Write(data); err != nil {
		return fmt.Errorf("failed to write HEIC image: %v", err)
	}

	return nil
}
