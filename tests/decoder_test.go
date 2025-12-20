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

func TestDecode(t *testing.T) {
	t.Run("successful decode", func(t *testing.T) {
		data, err := os.ReadFile("../assets/image.heic")
		require.NoError(t, err)

		reader := bytes.NewReader(data)
		img, err := webp.Decode(reader)

		assert.NoError(t, err)
		assert.NotNil(t, img)
		assert.Equal(t, img.Bounds().Dx(), 1024)
		assert.Equal(t, img.Bounds().Dy(), 1442)
	})

	t.Run("empty reader", func(t *testing.T) {
		reader := bytes.NewReader([]byte{})
		img, err := webp.Decode(reader)

		assert.Error(t, err)
		assert.Nil(t, img)
	})

	t.Run("invalid HEIF data", func(t *testing.T) {
		reader := bytes.NewReader([]byte("invalid heif data"))
		img, err := webp.Decode(reader)

		assert.Error(t, err)
		assert.Nil(t, img)
	})

	t.Run("reader error", func(t *testing.T) {
		reader := &errorReader{}
		img, err := webp.Decode(reader)

		assert.Error(t, err)
		assert.Nil(t, img)
		assert.Contains(t, err.Error(), "failed to decode HEIF data")
	})
}

func TestDecodeConfig(t *testing.T) {
	t.Run("successful decode config", func(t *testing.T) {
		data, err := os.ReadFile("../assets/image.heic")
		require.NoError(t, err)

		reader := bytes.NewReader(data)
		config, err := webp.DecodeConfig(reader)

		assert.NoError(t, err)
		assert.Equal(t, config.Width, 1024)
		assert.Equal(t, config.Height, 1442)
		assert.NotNil(t, config.ColorModel)
	})

	t.Run("empty reader", func(t *testing.T) {
		reader := bytes.NewReader([]byte{})
		config, err := webp.DecodeConfig(reader)

		assert.Error(t, err)
		assert.Equal(t, 0, config.Width)
		assert.Equal(t, 0, config.Height)
	})

	t.Run("invalid HEIF data", func(t *testing.T) {
		reader := bytes.NewReader([]byte("invalid heif data"))
		config, err := webp.DecodeConfig(reader)

		assert.Error(t, err)
		assert.Equal(t, 0, config.Width)
		assert.Equal(t, 0, config.Height)
	})

	t.Run("reader error", func(t *testing.T) {
		reader := &errorReader{}
		config, err := webp.DecodeConfig(reader)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed get config of HEIC data")
		assert.Equal(t, image.Config{}, config)
	})
}

func TestImageRegistration(t *testing.T) {
	t.Run("heic format registered", func(t *testing.T) {
		data, err := os.ReadFile("../assets/image.heic")
		require.NoError(t, err)

		reader := bytes.NewReader(data)
		_, format, err := image.Decode(reader)

		assert.NoError(t, err)
		assert.Equal(t, "heic", format)
	})

	t.Run("decode config through image package", func(t *testing.T) {
		data, err := os.ReadFile("../assets/image.heic")
		require.NoError(t, err)

		reader := bytes.NewReader(data)
		config, format, err := image.DecodeConfig(reader)

		assert.NoError(t, err)
		assert.Equal(t, "heic", format)
		assert.Equal(t, config.Width, 1024)
		assert.Equal(t, config.Height, 1442)
	})
}

func TestDecodeRoundtrip(t *testing.T) {
	t.Run("decode and verify image properties", func(t *testing.T) {
		data, err := os.ReadFile("../assets/image.heic")
		require.NoError(t, err)

		// First get config
		configReader := bytes.NewReader(data)
		config, err := webp.DecodeConfig(configReader)
		require.NoError(t, err)

		// Then decode the full image
		decodeReader := bytes.NewReader(data)
		img, err := webp.Decode(decodeReader)
		require.NoError(t, err)

		// Verify dimensions match
		assert.Equal(t, config.Width, img.Bounds().Dx())
		assert.Equal(t, config.Height, img.Bounds().Dy())

		// Verify the color model
		assert.Equal(t, config.ColorModel, color.RGBAModel)
	})
}

// errorReader is a mock reader that always returns an error
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}
