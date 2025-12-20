// Package webp is a Go library and CLI tool to encode/decode HEIF/HEIC images without system dependencies (CGO).
package webp

/*
#include <stdlib.h>
#include <string.h>
#include <libheif/heif.h>

// Memory writer structure to capture encoded HEIF data
typedef struct {
    uint8_t* data;
    size_t size;
    size_t capacity;
} memory_writer;

// Writer callback: appends data to our growing buffer
static struct heif_error writer_write(struct heif_context* ctx, const void* data, size_t size, void* userdata) {
    memory_writer* writer = (memory_writer*)userdata;

    // Grow buffer if needed
    if (writer->size + size > writer->capacity) {
        size_t new_capacity = writer->capacity * 2;
        if (new_capacity < writer->size + size) {
            new_capacity = writer->size + size;
        }
        uint8_t* new_data = (uint8_t*)realloc(writer->data, new_capacity);
        if (!new_data) {
            struct heif_error err = {heif_error_Memory_allocation_error, heif_suberror_Unspecified, "Out of memory"};
            return err;
        }
        writer->data = new_data;
        writer->capacity = new_capacity;
    }

    // Append data
    memcpy(writer->data + writer->size, data, size);
    writer->size += size;

    struct heif_error err = {heif_error_Ok, heif_suberror_Unspecified, "Success"};
    return err;
}

// Constants
#define INITIAL_BUFFER_SIZE (64 * 1024)
#define BYTES_PER_PIXEL 4

// Encode HEIF image to memory buffer
uint8_t* encode_heif_to_memory(struct heif_context* ctx, size_t* out_size, size_t estimated_size) {
    if (!ctx || !out_size) {
        if (out_size) *out_size = 0;
        return NULL;
    }

    memory_writer writer;
    // Use estimated size or default initial buffer size
    size_t initial_capacity = estimated_size > 0 && estimated_size < INITIAL_BUFFER_SIZE
                              ? estimated_size : INITIAL_BUFFER_SIZE;

    writer.data = (uint8_t*)malloc(initial_capacity);
    if (!writer.data) {
        *out_size = 0;
        return NULL;
    }
    writer.size = 0;
    writer.capacity = initial_capacity;

    struct heif_writer heif_writer;
    heif_writer.writer_api_version = 1;
    heif_writer.write = writer_write;

    struct heif_error err = heif_context_write(ctx, &heif_writer, &writer);
    if (err.code != heif_error_Ok) {
        free(writer.data);
        *out_size = 0;
        return NULL;
    }

    *out_size = writer.size;
    return writer.data;
}

// Full decode: reads HEIF data from memory, gets the primary image,
// decodes it into an interleaved RGBA plane, and returns the heif_image*.
// Also returns the heif_context* and heif_image_handle* for cleanup.
struct heif_image* decode_heif_image(const uint8_t *data, size_t size,
                              struct heif_context **outCtx,
                              struct heif_image_handle **outHandle) {
    // Validate input parameters
    if (!data || size == 0) {
        return NULL;
    }

    struct heif_context* ctx = heif_context_alloc();
    if (!ctx) return NULL;

    struct heif_error err = heif_context_read_from_memory(ctx, data, size, NULL);
    if (err.code != heif_error_Ok) {
        heif_context_free(ctx);
        return NULL;
    }

    struct heif_image_handle* handle = NULL;
    err = heif_context_get_primary_image_handle(ctx, &handle);
    if (err.code != heif_error_Ok) {
        heif_context_free(ctx);
        return NULL;
    }

    struct heif_image* img = NULL;
    // ask for interleaved RGBA
    err = heif_decode_image(handle, &img,
                            heif_colorspace_RGB,
                            heif_chroma_interleaved_RGBA,
                            NULL);
    if (err.code != heif_error_Ok) {
        heif_image_handle_release(handle);
        heif_context_free(ctx);
        return NULL;
    }

    if (outCtx)    *outCtx    = ctx;
    if (outHandle) *outHandle = handle;
    return img;
}

// get_heif_config: reads just enough of the HEIF file to extract width/height.
void get_heif_config(const uint8_t *data, size_t size,
                     uint32_t *width, uint32_t *height) {
    // Validate input parameters
    if (!data || size == 0 || !width || !height) {
        if (width) *width = 0;
        if (height) *height = 0;
        return;
    }

    struct heif_context* ctx = heif_context_alloc();
    if (!ctx) {
        *width = 0;
        *height = 0;
        return;
    }

    struct heif_error err = heif_context_read_from_memory(ctx, data, size, NULL);
    if (err.code != heif_error_Ok) {
        *width = 0;
        *height = 0;
        heif_context_free(ctx);
        return;
    }

    struct heif_image_handle* handle = NULL;
    err = heif_context_get_primary_image_handle(ctx, &handle);
    if (err.code != heif_error_Ok) {
        *width = 0;
        *height = 0;
        heif_context_free(ctx);
        return;
    }

    *width  = (uint32_t)heif_image_handle_get_width(handle);
    *height = (uint32_t)heif_image_handle_get_height(handle);

    heif_image_handle_release(handle);
    heif_context_free(ctx);
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

func encodeHEIF(rgba image.RGBA, options Options) ([]byte, error) {
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

	// Create the libheif context
	ctx := C.heif_context_alloc()
	if ctx == nil {
		return nil, fmt.Errorf("failed to allocate HEIF context")
	}
	defer C.heif_context_free(ctx)

	// Create an heicImage for the output
	var heicImage *C.struct_heif_image
	errCreate := C.heif_image_create(C.int(width), C.int(height), C.heif_colorspace_RGB, C.heif_chroma_interleaved_RGBA,
		&heicImage)

	if errCreate.code != C.heif_error_Ok {
		return nil, fmt.Errorf("failed to create HEIC image: %v", C.GoString(errCreate.message))
	}
	defer C.heif_image_release(heicImage)

	// Allocate the RGBA plane (8 bits)
	errPlane := C.heif_image_add_plane(heicImage, C.heif_channel_interleaved, C.int(width), C.int(height), C.int(8))
	if errPlane.code != C.heif_error_Ok {
		return nil, fmt.Errorf("failed to add RGBA plane to HEIC image: %v", C.GoString(errPlane.message))
	}

	// Copy the pixels
	var stride C.int
	ptr := C.heif_image_get_plane(heicImage, C.heif_channel_interleaved, &stride)
	planeSize := C.size_t(stride) * C.size_t(height)
	C.memcpy(unsafe.Pointer(ptr), unsafe.Pointer(&rgba.Pix[0]), planeSize)

	// Pick & configure HEVC encoder
	var encoder *C.struct_heif_encoder
	errEnc := C.heif_context_get_encoder_for_format(ctx, C.heif_compression_HEVC, &encoder)
	if errEnc.code != C.heif_error_Ok {
		return nil, fmt.Errorf("failed to create HEIC encoder: %v", C.GoString(errEnc.message))
	}
	defer C.heif_encoder_release(encoder)

	// Set the image quality
	var errQ C.struct_heif_error
	if options.Quality < losslessQuality {
		errQ = C.heif_encoder_set_lossy_quality(encoder, C.int(options.Quality))
	} else {
		errQ = C.heif_encoder_set_lossless(encoder, C.int(1))
	}

	if errQ.code != C.heif_error_Ok {
		return nil, fmt.Errorf("failed to set the image quality: %v", C.GoString(errQ.message))
	}

	// Encode into the context
	var handle *C.struct_heif_image_handle
	errImg := C.heif_context_encode_image(ctx, heicImage, encoder, nil, &handle)
	if errImg.code != C.heif_error_Ok {
		return nil, fmt.Errorf("failed to encode HEIC image: %v", C.GoString(errImg.message))
	}
	defer C.heif_image_handle_release(handle)

	// Encode to memory directly with size estimate
	estimatedSize := C.size_t(width * height * bytesPerPixel / 10) // rough estimate: 10% of raw size
	var size C.size_t
	cData := C.encode_heif_to_memory(ctx, &size, estimatedSize)
	if cData == nil {
		return nil, fmt.Errorf("failed to encode HEIF image to memory")
	}
	defer C.free(unsafe.Pointer(cData))

	// Copy C memory to Go slice
	data := C.GoBytes(unsafe.Pointer(cData), C.int(size))

	return data, nil
}

func decodeHEIFToRGBA(data []byte) (*image.RGBA, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data buffer")
	}

	// Pin the data to prevent GC relocation
	cData := C.CBytes(data)
	defer C.free(cData)

	// Call our C helper
	var ctx *C.struct_heif_context
	var handle *C.struct_heif_image_handle
	img := C.decode_heif_image((*C.uint8_t)(cData), C.size_t(len(data)), &ctx, &handle)
	if img == nil {
		return nil, fmt.Errorf("failed to decode HEIF image")
	}
	defer C.heif_image_release(img)
	defer C.heif_image_handle_release(handle)
	defer C.heif_context_free(ctx)

	// Query width/height from the interleaved plane
	width := int(C.heif_image_get_width(img, C.heif_channel_interleaved))
	height := int(C.heif_image_get_height(img, C.heif_channel_interleaved))

	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid decoded image dimensions: %dx%d", width, height)
	}

	// Grab a pointer to the RGBA data and its stride
	var cStride C.int
	ptr := C.heif_image_get_plane_readonly(img, C.heif_channel_interleaved, &cStride)
	rowBytes := int(cStride)

	// Allocate our Go RGBA
	goImg := image.NewRGBA(image.Rect(0, 0, width, height))

	// Direct memory copy - more efficient than row-by-row with intermediate allocation
	rowSize := width * bytesPerPixel
	for y := 0; y < height; y++ {
		srcPtr := unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + uintptr(y*rowBytes))
		dstOff := y * goImg.Stride
		// Direct unsafe copy using unsafe.Slice
		srcSlice := unsafe.Slice((*byte)(srcPtr), rowSize)
		copy(goImg.Pix[dstOff:dstOff+rowSize], srcSlice)
	}

	return goImg, nil
}

// DecodeConfig reads enough of data to determine the image's configuration (dimensions, etc.).
// Here we read the entire data and call a lightweight C function that only parses the header.
func decodeConfig(data []byte) (image.Config, error) {
	if len(data) == 0 {
		return image.Config{}, fmt.Errorf("empty data buffer")
	}

	// Pin the data to prevent GC relocation
	cData := C.CBytes(data)
	defer C.free(cData)

	var w, h C.uint32_t
	C.get_heif_config(
		(*C.uint8_t)(cData),
		C.size_t(len(data)),
		&w,
		&h,
	)

	if w == 0 || h == 0 {
		return image.Config{}, fmt.Errorf("failed to get HEIF image config")
	}

	return image.Config{
		ColorModel: color.RGBAModel,
		Width:      int(w),
		Height:     int(h),
	}, nil
}
