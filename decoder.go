//go:build cgo

package webp

import (
	"fmt"
	"image"
	"io"
)

// The init function registers the WebP decoder with Go's image package.
// WebP files start with "RIFF" followed by file size and "WEBP" signature
func init() {
	image.RegisterFormat("webp", "RIFF????WEBP", Decode, DecodeConfig)
}

// Decode reads WebP image data from the provided io.Reader and decodes it into an image.Image.
//
// It returns the decoded image or an error if the decoding process fails.
func Decode(reader io.Reader) (image.Image, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode WebP data: %w", err)
	}
	return decodeWebPToRGBA(data)
}

// DecodeConfig reads the configuration of a WebP image from the provided io.Reader.
//
// It returns an image.Config containing the width, height, and color model of the image, or an error if the
// configuration cannot be determined.
func DecodeConfig(reader io.Reader) (image.Config, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return image.Config{}, fmt.Errorf("failed get config of WebP data: %w", err)
	}

	return decodeConfig(data)
}
