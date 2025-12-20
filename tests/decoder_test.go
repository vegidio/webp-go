//go:build cgo

package tests

import (
	"bytes"
	"image"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vegidio/webp-go"
)

func TestDecode(t *testing.T) {
	t.Run("ValidWebP", func(t *testing.T) {
		data, err := os.ReadFile("../assets/image.webp")
		require.NoError(t, err, "Failed to read test WebP file")

		reader := bytes.NewReader(data)
		img, err := webp.Decode(reader)

		assert.NoError(t, err)
		assert.NotNil(t, img)
		assert.Greater(t, img.Bounds().Dx(), 0)
		assert.Greater(t, img.Bounds().Dy(), 0)
	})

	t.Run("InvalidData", func(t *testing.T) {
		invalidData := []byte("not a valid webp file")
		reader := bytes.NewReader(invalidData)

		img, err := webp.Decode(reader)

		assert.Error(t, err)
		assert.Nil(t, img)
	})

	t.Run("EmptyReader", func(t *testing.T) {
		reader := bytes.NewReader([]byte{})

		img, err := webp.Decode(reader)

		assert.Error(t, err)
		assert.Nil(t, img)
	})

	t.Run("ReadError", func(t *testing.T) {
		reader := &errorReader{}

		img, err := webp.Decode(reader)

		assert.Error(t, err)
		assert.Nil(t, img)
		assert.Contains(t, err.Error(), "failed to decode WebP data")
	})
}

func TestDecodeConfig(t *testing.T) {
	t.Run("ValidWebP", func(t *testing.T) {
		data, err := os.ReadFile("../assets/image.webp")
		require.NoError(t, err, "Failed to read test WebP file")

		reader := bytes.NewReader(data)
		config, err := webp.DecodeConfig(reader)

		assert.NoError(t, err)
		assert.Greater(t, config.Width, 0)
		assert.Greater(t, config.Height, 0)
		assert.NotNil(t, config.ColorModel)
	})

	t.Run("InvalidData", func(t *testing.T) {
		invalidData := []byte("not a valid webp file")
		reader := bytes.NewReader(invalidData)

		config, err := webp.DecodeConfig(reader)

		assert.Error(t, err)
		assert.Equal(t, 0, config.Width)
		assert.Equal(t, 0, config.Height)
	})

	t.Run("EmptyReader", func(t *testing.T) {
		reader := bytes.NewReader([]byte{})

		config, err := webp.DecodeConfig(reader)

		assert.Error(t, err)
		assert.Equal(t, 0, config.Width)
		assert.Equal(t, 0, config.Height)
	})

	t.Run("ReadError", func(t *testing.T) {
		reader := &errorReader{}

		config, err := webp.DecodeConfig(reader)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed get config of WebP data")
		assert.Equal(t, 0, config.Width)
		assert.Equal(t, 0, config.Height)
	})
}

func TestWebPFormatRegistration(t *testing.T) {
	t.Run("DecodeWithImagePackage", func(t *testing.T) {
		data, err := os.ReadFile("../assets/image.webp")
		require.NoError(t, err, "Failed to read test WebP file")

		reader := bytes.NewReader(data)
		_, format, err := image.Decode(reader)

		assert.NoError(t, err)
		assert.Equal(t, "webp", format)
	})

	t.Run("DecodeConfigWithImagePackage", func(t *testing.T) {
		data, err := os.ReadFile("../assets/image.webp")
		require.NoError(t, err, "Failed to read test WebP file")

		reader := bytes.NewReader(data)
		config, format, err := image.DecodeConfig(reader)

		assert.NoError(t, err)
		assert.Equal(t, "webp", format)
		assert.Greater(t, config.Width, 0)
		assert.Greater(t, config.Height, 0)
	})
}

// errorReader is a mock reader that always returns an error
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}
