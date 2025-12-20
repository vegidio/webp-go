//go:build cgo

package webp

import (
	"fmt"
	"image"
	"io"
)

// The init function registers the HEIF decoder with Go's image package.
// The second argument ("????ftypheic", "????ftypheix", etc) lists substrings expected in the file header.
func init() {
	image.RegisterFormat("heic", "????ftypheic", Decode, DecodeConfig)
	image.RegisterFormat("heic", "????ftypheix", Decode, DecodeConfig)
	image.RegisterFormat("heic", "????ftyphev1", Decode, DecodeConfig)
	image.RegisterFormat("heic", "????ftyphevx", Decode, DecodeConfig)

}

// Decode reads HEIF image data from the provided io.Reader and decodes it into an image.Image.
//
// It returns the decoded image or an error if the decoding process fails.
func Decode(reader io.Reader) (image.Image, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode HEIF data: %w", err)
	}
	return decodeHEIFToRGBA(data)
}

// DecodeConfig reads the configuration of a HEIF image from the provided io.Reader.
//
// It returns an image.Config containing the width, height, and color model of the image, or an error if the
// configuration cannot be determined.
func DecodeConfig(reader io.Reader) (image.Config, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return image.Config{}, fmt.Errorf("failed get config of HEIC data: %w", err)
	}

	return decodeConfig(data)
}
