package mtx

import (
	"reflect"
)

var (
	FILE_HEADER_SIZE = int(reflect.TypeOf(FileHeader{}).Size())
	IMAGE_HEADER_SIZE = int(reflect.TypeOf(ImageHeader{}).Size())
)

const (
	MAX_IMAGE_SIZE = 4096
)

type FileHeader struct {
	Magic uint32
	SecondImageOffset uint32
	SizeCheck uint32
}

type ImageHeader struct {
	Magic uint32
	Width uint32
	Height uint32
}