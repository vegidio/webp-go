//go:build cgo

package tests

import (
	"bytes"
	"image"
	"image/color"
	_ "image/jpeg"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vegidio/webp-go"
)

func TestEncode(t *testing.T) {
	t.Run("ValidEncodeWithDefaultOptions", func(t *testing.T) {
		img := createTestImage(100, 100)
		buf := &bytes.Buffer{}

		err := webp.Encode(buf, img, nil)

		assert.NoError(t, err)
		assert.InDelta(t, 402, buf.Len(), 50, "Encoded size should be close to expected")
	})

	t.Run("ValidEncodeWithCustomQuality", func(t *testing.T) {
		img := createTestImage(100, 100)
		buf := &bytes.Buffer{}
		options := &webp.Options{Quality: 90}

		err := webp.Encode(buf, img, options)

		assert.NoError(t, err)
		assert.InDelta(t, 620, buf.Len(), 60, "Encoded size should be close to expected")
	})

	t.Run("ValidEncodeWithMinQuality", func(t *testing.T) {
		img := createTestImage(100, 100)
		buf := &bytes.Buffer{}
		options := &webp.Options{Quality: 0}

		err := webp.Encode(buf, img, options)

		assert.NoError(t, err)
		assert.InDelta(t, 192, buf.Len(), 30, "Encoded size should be close to expected")
	})

	t.Run("LossyQualityProgression", func(t *testing.T) {
		// Test that higher lossy quality (0-99) generally produces larger files
		img := createTestImage(100, 100)

		bufLow := &bytes.Buffer{}
		err := webp.Encode(bufLow, img, &webp.Options{Quality: 25})
		assert.NoError(t, err)

		bufHigh := &bytes.Buffer{}
		err = webp.Encode(bufHigh, img, &webp.Options{Quality: 90})
		assert.NoError(t, err)

		assert.Greater(t, bufHigh.Len(), bufLow.Len(), "Higher quality should produce larger file in lossy mode")
	})

	t.Run("LosslessEncoding", func(t *testing.T) {
		// Quality 100 uses lossless compression, which may be smaller or larger than lossy
		// depending on image content. The key is that it preserves all pixels exactly.
		img := createTestImage(100, 100)
		buf := &bytes.Buffer{}
		options := &webp.Options{Quality: 100}

		err := webp.Encode(buf, img, options)
		assert.NoError(t, err)
		assert.InDelta(t, 102, buf.Len(), 20, "Encoded size should be close to expected")

		// Verify lossless encoding preserves pixels exactly
		decoded, err := webp.Decode(bytes.NewReader(buf.Bytes()))
		assert.NoError(t, err)

		// Check a few pixels to verify they match exactly
		for _, point := range []image.Point{{0, 0}, {50, 50}, {99, 99}} {
			original := img.At(point.X, point.Y)
			decodedColor := decoded.At(point.X, point.Y)
			assert.Equal(t, original, decodedColor, "Lossless encoding should preserve pixel at (%d,%d)", point.X, point.Y)
		}
	})

	t.Run("InvalidQualityTooLow", func(t *testing.T) {
		img := createTestImage(100, 100)
		buf := &bytes.Buffer{}
		options := &webp.Options{Quality: -1}

		err := webp.Encode(buf, img, options)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "quality must be between 0 and 100")
	})

	t.Run("InvalidQualityTooHigh", func(t *testing.T) {
		img := createTestImage(100, 100)
		buf := &bytes.Buffer{}
		options := &webp.Options{Quality: 101}

		err := webp.Encode(buf, img, options)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "quality must be between 0 and 100")
	})

	t.Run("EncodeNRGBAImage", func(t *testing.T) {
		nrgba := image.NewNRGBA(image.Rect(0, 0, 50, 50))
		for y := 0; y < 50; y++ {
			for x := 0; x < 50; x++ {
				nrgba.Set(x, y, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
			}
		}
		buf := &bytes.Buffer{}

		err := webp.Encode(buf, nrgba, nil)

		assert.NoError(t, err)
		assert.InDelta(t, 98, buf.Len(), 20, "Encoded size should be close to expected")
	})

	t.Run("EncodeGrayImage", func(t *testing.T) {
		gray := image.NewGray(image.Rect(0, 0, 50, 50))
		for y := 0; y < 50; y++ {
			for x := 0; x < 50; x++ {
				gray.Set(x, y, color.Gray{Y: 128})
			}
		}
		buf := &bytes.Buffer{}

		err := webp.Encode(buf, gray, nil)

		assert.NoError(t, err)
		assert.InDelta(t, 64, buf.Len(), 15, "Encoded size should be close to expected")
	})

	t.Run("WriteError", func(t *testing.T) {
		img := createTestImage(100, 100)
		writer := &errorWriter{}

		err := webp.Encode(writer, img, nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write WebP image")
	})

	t.Run("EncodeJPGFile", func(t *testing.T) {
		data, err := os.ReadFile("../assets/image.jpg")
		require.NoError(t, err, "Failed to read test JPG file")

		img, _, err := image.Decode(bytes.NewReader(data))
		require.NoError(t, err, "Failed to decode JPG file")

		buf := &bytes.Buffer{}
		err = webp.Encode(buf, img, &webp.Options{Quality: 80})

		assert.NoError(t, err)
		assert.InDelta(t, 95_706, buf.Len(), 1_000, "Encoded size should be close to expected")
	})

	t.Run("EncodeWebPFile", func(t *testing.T) {
		data, err := os.ReadFile("../assets/image.webp")
		require.NoError(t, err, "Failed to read test WebP file")

		img, err := webp.Decode(bytes.NewReader(data))
		require.NoError(t, err, "Failed to decode WebP file")

		buf := &bytes.Buffer{}
		err = webp.Encode(buf, img, &webp.Options{Quality: 85})

		assert.NoError(t, err)
		assert.InDelta(t, 87_194, buf.Len(), 1_000, "Encoded size should be close to expected")
	})

	t.Run("EncodeSmallImage", func(t *testing.T) {
		img := createTestImage(1, 1)
		buf := &bytes.Buffer{}

		err := webp.Encode(buf, img, nil)

		assert.NoError(t, err)
		assert.InDelta(t, 62, buf.Len(), 15, "Encoded size should be close to expected")
	})

	t.Run("EncodeLargeImage", func(t *testing.T) {
		img := createTestImage(1000, 1000)
		buf := &bytes.Buffer{}

		err := webp.Encode(buf, img, &webp.Options{Quality: 50})

		assert.NoError(t, err)
		assert.InDelta(t, 5536, buf.Len(), 300, "Encoded size should be close to expected")
	})
}

// createTestImage creates a simple test image with the given dimensions
func createTestImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with a gradient pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8((x * 255) / width)
			g := uint8((y * 255) / height)
			b := uint8(128)
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	return img
}

// errorWriter is a mock writer that always returns an error
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}
