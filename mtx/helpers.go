package mtx

import (
	"bytes"
	"compress/zlib"
	"image"
	"image/draw"
	"io"
	"os"
)

func readSomeBytes(file *os.File, number int) ([]byte, error) {
	b := make([]byte, number)

	_, err := file.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func decompressZlibData(data []byte) ([]byte, error) {
	b := bytes.NewReader(data)
	z, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}
	defer z.Close()

	decompBytes, err := io.ReadAll(z)
	if err != nil {
		return nil, err
	}

	return decompBytes, nil
}

func compressZlibData(data []byte) ([]byte, error) {
	var b bytes.Buffer
	z, err := zlib.NewWriterLevel(&b, zlib.BestCompression)
	if err != nil {
		return nil, err
	}

	_, err = z.Write(data)
	z.Close()
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func newGrayFromRawData(data []byte, width int, height int) *image.Gray {
	img := image.NewGray(image.Rectangle{
		Min: image.Point{},
		Max: image.Point{X: width, Y: height},
	})
	img.Pix = data
	return img
}

func imageToNRGBA(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	nrgba := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(nrgba, bounds, img, bounds.Min, draw.Src)
	return nrgba
}

func getAlphaChannel(img *image.NRGBA) []byte {
	capacity := len(img.Pix) / 4
	alpha := make([]byte, capacity)

	for i := 0; i < capacity; i++ {
		alphaIdx := i*4 + 3
		alpha[i] = img.Pix[alphaIdx]
	}

	return alpha
}

func makeAlphaChannelOpaque(rgba *image.NRGBA) {
	capacity := len(rgba.Pix) / 4

	for i := 0; i < capacity; i++ {
		alphaIdx := i*4 + 3
		rgba.Pix[alphaIdx] = 0xFF
	}
}
