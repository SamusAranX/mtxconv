package mtx

import (
	"bytes"
	"compress/zlib"
	"image"
	"image/draw"
	"io/ioutil"
	"os"
)

func readSomeBytes(file *os.File, number int) ([]byte, error) {
	bytes := make([]byte, number)

	_, err := file.Read(bytes)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func decompressZlibData(data []byte) ([]byte, error) {
	b := bytes.NewReader(data)
	z, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}
	defer z.Close()

	decompBytes, err := ioutil.ReadAll(z)
	if err != nil {
		return nil, err
	}

	return decompBytes, nil
}

func compressZlibData(data []byte) ([]byte, error) {
	var b bytes.Buffer
	z := zlib.NewWriter(&b)
	defer z.Close()

	_, err := z.Write(data)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func imageToNRGBA(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	nrgba := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(nrgba, bounds, img, bounds.Min, draw.Src)
	return nrgba
}

func newGrayFromRawData(data []byte, width int, height int) *image.Gray {
	img := image.NewGray(image.Rectangle{
		Min: image.Point{},
		Max: image.Point{X: width, Y: height},
	})
	img.Pix = data
	return img
}
