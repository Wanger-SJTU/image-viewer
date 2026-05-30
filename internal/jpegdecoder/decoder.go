package jpegdecoder

/*
#cgo LDFLAGS: -ljpeg
#include <stdio.h>
#include <stdlib.h>
#include <jpeglib.h>
#include <setjmp.h>

struct my_error_mgr {
	struct jpeg_error_mgr pub;
	jmp_buf setjmp_buffer;
};

static void my_error_exit(j_common_ptr cinfo) {
	struct my_error_mgr *mgr = (struct my_error_mgr *)cinfo->err;
	(*cinfo->err->output_message)(cinfo);
	longjmp(mgr->setjmp_buffer, 1);
}

// read_jpeg_dimensions quickly reads image dimensions from a JPEG header without
// decompressing. Returns 0 on success, -1 on error.
static int read_jpeg_dimensions(const char *filename, int *width, int *height) {
	FILE *fp = fopen(filename, "rb");
	if (!fp) return -1;

	struct jpeg_decompress_struct cinfo;
	struct my_error_mgr jerr;
	cinfo.err = jpeg_std_error(&jerr.pub);
	jerr.pub.error_exit = my_error_exit;

	if (setjmp(jerr.setjmp_buffer)) {
		jpeg_destroy_decompress(&cinfo);
		fclose(fp);
		return -1;
	}

	jpeg_create_decompress(&cinfo);
	jpeg_stdio_src(&cinfo, fp);
	jpeg_read_header(&cinfo, TRUE);

	*width = cinfo.image_width;
	*height = cinfo.image_height;

	jpeg_destroy_decompress(&cinfo);
	fclose(fp);
	return 0;
}

// decode_jpeg opens a file, decompresses with libjpeg at the given DCT scale,
// and returns raw pixel data. scale_num/denom control output resolution:
// 1/1 = full, 1/2 = half, 1/4 = quarter, 1/8 = eighth.
// Caller must free the returned buffer.
static unsigned char* decode_jpeg_file(
	const char *filename,
	int *width,
	int *height,
	int *channels,
	int scale_num,
	int scale_denom,
	char **errmsg
) {
	FILE *fp = fopen(filename, "rb");
	if (!fp) {
		*errmsg = "cannot open file";
		return NULL;
	}

	struct jpeg_decompress_struct cinfo;
	struct my_error_mgr jerr;

	cinfo.err = jpeg_std_error(&jerr.pub);
	jerr.pub.error_exit = my_error_exit;

	if (setjmp(jerr.setjmp_buffer)) {
		jpeg_destroy_decompress(&cinfo);
		fclose(fp);
		*errmsg = "jpeg decode failed";
		return NULL;
	}

	jpeg_create_decompress(&cinfo);
	jpeg_stdio_src(&cinfo, fp);
	jpeg_read_header(&cinfo, TRUE);

	// DCT-domain downscaling — discards high-freq coefficients during decode
	cinfo.scale_num = scale_num;
	cinfo.scale_denom = scale_denom;

	jpeg_start_decompress(&cinfo);

	*width  = cinfo.output_width;
	*height = cinfo.output_height;
	*channels = cinfo.output_components;

	unsigned long row_stride = (*width) * (*channels);
	unsigned long data_size = row_stride * (*height);
	unsigned char *data = (unsigned char *)malloc(data_size);
	if (!data) {
		jpeg_destroy_decompress(&cinfo);
		fclose(fp);
		*errmsg = "malloc failed";
		return NULL;
	}

	while (cinfo.output_scanline < cinfo.output_height) {
		unsigned char *row_ptr = data + cinfo.output_scanline * row_stride;
		JSAMPROW row_ptrs[1] = { row_ptr };
		jpeg_read_scanlines(&cinfo, row_ptrs, 1);
	}

	jpeg_finish_decompress(&cinfo);
	jpeg_destroy_decompress(&cinfo);
	fclose(fp);

	*errmsg = NULL;
	return data;
}
*/
import "C"

import (
	"fmt"
	"image"
	"image/color"
	"unsafe"
)

// ReadDimensions reads the image dimensions from a JPEG header without decompressing.
func ReadDimensions(path string) (int, int, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var width, height C.int
	if C.read_jpeg_dimensions(cPath, &width, &height) != 0 {
		return 0, 0, fmt.Errorf("jpeg read header: failed")
	}
	return int(width), int(height), nil
}

// DecodeFile decodes a JPEG file at full resolution using libjpeg-turbo.
func DecodeFile(path string) (image.Image, error) {
	return decodeFileScaled(path, 1, 1)
}

// DecodeFileScaled decodes a JPEG at reduced resolution using DCT-domain scaling.
// scale is the denominator: 1=full, 2=half, 4=quarter, 8=eighth.
func DecodeFileScaled(path string, scale int) (image.Image, error) {
	if scale < 1 {
		scale = 1
	}
	if scale > 8 {
		scale = 8
	}
	// Normalize: libjpeg only supports 1/1, 1/2, 1/4, 1/8
	switch {
	case scale >= 8:
		return decodeFileScaled(path, 1, 8)
	case scale >= 4:
		return decodeFileScaled(path, 1, 4)
	case scale >= 2:
		return decodeFileScaled(path, 1, 2)
	default:
		return decodeFileScaled(path, 1, 1)
	}
}

func decodeFileScaled(path string, scaleNum, scaleDenom int) (image.Image, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var width, height, channels C.int
	var errmsg *C.char

	data := C.decode_jpeg_file(cPath, &width, &height, &channels, C.int(scaleNum), C.int(scaleDenom), &errmsg)
	if data == nil {
		msg := C.GoString(errmsg)
		return nil, fmt.Errorf("jpeg decode: %s", msg)
	}
	defer C.free(unsafe.Pointer(data))

	w := int(width)
	h := int(height)
	c := int(channels)

	length := w * h * c
	raw := (*[1 << 30]C.uchar)(unsafe.Pointer(data))[:length:length]

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	switch c {
	case 3:
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				idx := (y*w + x) * 3
				img.Set(x, y, color.RGBA{
					R: uint8(raw[idx]),
					G: uint8(raw[idx+1]),
					B: uint8(raw[idx+2]),
					A: 255,
				})
			}
		}
	case 1:
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				idx := y*w + x
				v := uint8(raw[idx])
				img.Set(x, y, color.RGBA{R: v, G: v, B: v, A: 255})
			}
		}
	default:
		return nil, fmt.Errorf("unsupported number of channels: %d", c)
	}

	return img, nil
}
