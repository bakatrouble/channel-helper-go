package utils

/*
#cgo LDFLAGS: -L../lib -limagehash
#include <../lib/imagehash.h>
#include <stdlib.h>
*/
import "C"
import (
	"bytes"
	"image/jpeg"
	"unsafe"
)

func HashImage(imageData []byte) (string, error) {
	// first try to open the image to verify it's a valid image
	_, err := jpeg.Decode(bytes.NewReader(imageData))
	if err != nil {
		return "", err
	}

	imageDataPtr := unsafe.Pointer(&imageData[0])
	imageDataLen := len(imageData)

	cImageHash := C.hash_image((*C.uchar)(imageDataPtr), C.uint32_t(imageDataLen))
	defer C.free(unsafe.Pointer(cImageHash))
	if cImageHash == nil {
		return "", nil
	}

	return C.GoString(cImageHash), nil
}
