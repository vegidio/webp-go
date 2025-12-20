// Package webp is a Go library and CLI tool to encode/decode WebP images without system dependencies (CGO).
package webp

/*
#include <stdlib.h>
#include <string.h>
#include <encode.h>
#include <decode.h>

// Constants
#define BYTES_PER_PIXEL 4

// Encode WebP image to memory buffer using lossy compression
uint8_t* encode_webp_lossy(const uint8_t* rgba, int width, int height, int stride,
                            float quality, size_t* out_size) {
    if (!rgba || width <= 0 || height <= 0 || !out_size) {
        if (out_size) *out_size = 0;
        return NULL;
    }

    uint8_t* output = NULL;
    *out_size = WebPEncodeRGBA(rgba, width, height, stride, quality, &output);

    if (*out_size == 0) {
        return NULL;
    }

    return output;
}

// Encode WebP image to memory buffer using lossless compression
uint8_t* encode_webp_lossless(const uint8_t* rgba, int width, int height, int stride,
                               size_t* out_size) {
    if (!rgba || width <= 0 || height <= 0 || !out_size) {
        if (out_size) *out_size = 0;
        return NULL;
    }

    uint8_t* output = NULL;
    *out_size = WebPEncodeLosslessRGBA(rgba, width, height, stride, &output);

    if (*out_size == 0) {
        return NULL;
    }

    return output;
}

// Decode WebP image from memory buffer
uint8_t* decode_webp_to_rgba(const uint8_t* data, size_t data_size,
                              int* width, int* height) {
    if (!data || data_size == 0 || !width || !height) {
        if (width) *width = 0;
        if (height) *height = 0;
        return NULL;
    }

    uint8_t* output = WebPDecodeRGBA(data, data_size, width, height);

    if (!output || *width <= 0 || *height <= 0) {
        if (output) free(output);
        *width = 0;
        *height = 0;
        return NULL;
    }

    return output;
}

// Get WebP image configuration (dimensions)
int get_webp_config(const uint8_t* data, size_t data_size,
                    int* width, int* height) {
    if (!data || data_size == 0 || !width || !height) {
        if (width) *width = 0;
        if (height) *height = 0;
        return 0;
    }

    WebPBitstreamFeatures features;
    VP8StatusCode status = WebPGetFeatures(data, data_size, &features);

    if (status != VP8_STATUS_OK) {
        *width = 0;
        *height = 0;
        return 0;
    }

    *width = features.width;
    *height = features.height;
    return 1;
}
*/
import "C"

import (
	"fmt"
	"image"
	"image/color"
	"unsafe"
)

const (
	minQuality      = 0
	maxQuality      = 100
	bytesPerPixel   = 4
	losslessQuality = 100
)

func encodeWebP(rgba image.RGBA, options Options) ([]byte, error) {
	width := rgba.Bounds().Dx()
	height := rgba.Bounds().Dy()

	// Validate dimensions
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid image dimensions: %dx%d", width, height)
	}

	// Validate quality
	if options.Quality < minQuality || options.Quality > maxQuality {
		return nil, fmt.Errorf("quality must be between %d and %d, got %d", minQuality, maxQuality, options.Quality)
	}

	var cData *C.uint8_t
	var size C.size_t

	// Encode based on quality setting
	if options.Quality < losslessQuality {
		// Lossy compression
		cData = C.encode_webp_lossy(
			(*C.uint8_t)(unsafe.Pointer(&rgba.Pix[0])),
			C.int(width),
			C.int(height),
			C.int(rgba.Stride),
			C.float(options.Quality),
			&size,
		)
	} else {
		// Lossless compression
		cData = C.encode_webp_lossless(
			(*C.uint8_t)(unsafe.Pointer(&rgba.Pix[0])),
			C.int(width),
			C.int(height),
			C.int(rgba.Stride),
			&size,
		)
	}

	if cData == nil || size == 0 {
		return nil, fmt.Errorf("failed to encode WebP image")
	}
	defer C.WebPFree(unsafe.Pointer(cData))

	// Copy C memory to Go slice
	data := C.GoBytes(unsafe.Pointer(cData), C.int(size))

	return data, nil
}

func decodeWebPToRGBA(data []byte) (*image.RGBA, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data buffer")
	}

	var width, height C.int

	// Decode the WebP image
	cData := C.decode_webp_to_rgba(
		(*C.uint8_t)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
		&width,
		&height,
	)

	if cData == nil || width <= 0 || height <= 0 {
		return nil, fmt.Errorf("failed to decode WebP image")
	}
	defer C.WebPFree(unsafe.Pointer(cData))

	w := int(width)
	h := int(height)

	// Create a Go image
	goImg := image.NewRGBA(image.Rect(0, 0, w, h))

	// Copy decoded data to Go image
	dataSize := w * h * bytesPerPixel
	srcSlice := unsafe.Slice((*byte)(unsafe.Pointer(cData)), dataSize)
	copy(goImg.Pix, srcSlice)

	return goImg, nil
}

// decodeConfig reads enough of data to determine the image's configuration (dimensions, etc.).
func decodeConfig(data []byte) (image.Config, error) {
	if len(data) == 0 {
		return image.Config{}, fmt.Errorf("empty data buffer")
	}

	var width, height C.int
	result := C.get_webp_config(
		(*C.uint8_t)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
		&width,
		&height,
	)

	if result == 0 || width <= 0 || height <= 0 {
		return image.Config{}, fmt.Errorf("failed to get WebP image config")
	}

	return image.Config{
		ColorModel: color.RGBAModel,
		Width:      int(width),
		Height:     int(height),
	}, nil
}
