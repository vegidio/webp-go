//go:build cgo

package tests

import (
	"bytes"
	"image"
	"image/color"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vegidio/heif-go"
)

func TestEncode(t *testing.T) {
	t.Run("successful encode with default options", func(t *testing.T) {
		// Create a simple test image
		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		for y := 0; y < 100; y++ {
			for x := 0; x < 100; x++ {
				img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
			}
		}

		var buf bytes.Buffer
		err := webp.Encode(&buf, img, nil)

		assert.NoError(t, err)
		assert.Greater(t, buf.Len(), 0)
	})

	t.Run("successful encode with custom quality", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 50, 50))
		for y := 0; y < 50; y++ {
			for x := 0; x < 50; x++ {
				img.Set(x, y, color.RGBA{R: 0, G: 255, B: 0, A: 255})
			}
		}

		var buf bytes.Buffer
		options := &webp.Options{Quality: 80}
		err := webp.Encode(&buf, img, options)

		assert.NoError(t, err)
		assert.Greater(t, buf.Len(), 0)
	})

	t.Run("encode with quality 0", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 10, 10))

		var buf bytes.Buffer
		options := &webp.Options{Quality: 0}
		err := webp.Encode(&buf, img, options)

		assert.NoError(t, err)
		assert.Greater(t, buf.Len(), 0)
	})

	t.Run("encode with quality 100", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 10, 10))

		var buf bytes.Buffer
		options := &webp.Options{Quality: 100}
		err := webp.Encode(&buf, img, options)

		assert.NoError(t, err)
		assert.Greater(t, buf.Len(), 0)
	})

	t.Run("invalid quality less than 0", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 10, 10))

		var buf bytes.Buffer
		options := &webp.Options{Quality: -1}
		err := webp.Encode(&buf, img, options)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "quality must be between 0 and 100")
	})

	t.Run("invalid quality greater than 100", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 10, 10))

		var buf bytes.Buffer
		options := &webp.Options{Quality: 101}
		err := webp.Encode(&buf, img, options)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "quality must be between 0 and 100")
	})

	t.Run("encode different image type NRGBA", func(t *testing.T) {
		img := image.NewNRGBA(image.Rect(0, 0, 20, 20))
		for y := 0; y < 20; y++ {
			for x := 0; x < 20; x++ {
				img.Set(x, y, color.RGBA{R: 0, G: 0, B: 255, A: 255})
			}
		}

		var buf bytes.Buffer
		err := webp.Encode(&buf, img, &webp.Options{Quality: 70})

		assert.NoError(t, err)
		assert.Greater(t, buf.Len(), 0)
	})

	t.Run("writer error", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 10, 10))
		writer := &errorWriter{}

		err := webp.Encode(writer, img, nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write HEIC image")
	})
}

func TestEncodeRealImage(t *testing.T) {
	t.Run("encode decoded HEIC image", func(t *testing.T) {
		data, err := os.ReadFile("../assets/image.heic")
		require.NoError(t, err)

		reader := bytes.NewReader(data)
		img, err := webp.Decode(reader)
		require.NoError(t, err)

		var buf bytes.Buffer
		err = webp.Encode(&buf, img, &webp.Options{Quality: 75})

		assert.NoError(t, err)
		assert.Greater(t, buf.Len(), 0)

		// Verify the encoded image can be decoded
		decodeReader := bytes.NewReader(buf.Bytes())
		decodedImg, err := webp.Decode(decodeReader)
		assert.NoError(t, err)
		assert.NotNil(t, decodedImg)
		assert.Equal(t, img.Bounds().Dx(), decodedImg.Bounds().Dx())
		assert.Equal(t, img.Bounds().Dy(), decodedImg.Bounds().Dy())
	})
}

func TestEncodeRoundtrip(t *testing.T) {
	t.Run("encode and decode preserves dimensions", func(t *testing.T) {
		// Create original image
		width, height := 64, 48
		original := image.NewRGBA(image.Rect(0, 0, width, height))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				original.Set(x, y, color.RGBA{
					R: uint8(x * 255 / width),
					G: uint8(y * 255 / height),
					B: 128,
					A: 255,
				})
			}
		}

		// Encode
		var buf bytes.Buffer
		err := webp.Encode(&buf, original, &webp.Options{Quality: 90})
		require.NoError(t, err)

		// Decode
		reader := bytes.NewReader(buf.Bytes())
		decoded, err := webp.Decode(reader)
		require.NoError(t, err)

		// Verify dimensions
		assert.Equal(t, width, decoded.Bounds().Dx())
		assert.Equal(t, height, decoded.Bounds().Dy())
	})
}

// errorWriter is a mock writer that always returns an error
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, io.ErrShortWrite
}
