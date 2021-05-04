package mtx

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

func readSomeBytes(file *os.File, number int) []byte {
	bytes := make([]byte, number)

	_, err := file.Read(bytes)
	if err != nil {
		if err != io.EOF {
			log.Fatal(err)
		}
	}

	return bytes
}

func readFileHeader(file *os.File) (*FileHeader, error) {
	header := FileHeader{}
	headerData := readSomeBytes(file, int(FILE_HEADER_SIZE))
	headerBuf := bytes.NewBuffer(headerData)
	err := binary.Read(headerBuf, binary.LittleEndian, &header)
	if err != nil {
		return nil, err
	}

	return &header, nil
}

func readImageHeader(file *os.File) (*ImageHeader, error) {
	header := ImageHeader{}
	headerData := readSomeBytes(file, int(IMAGE_HEADER_SIZE))
	headerBuf := bytes.NewBuffer(headerData)
	err := binary.Read(headerBuf, binary.LittleEndian, &header)
	if err != nil {
		return nil, err
	}

	return &header, nil
}

func readDataChunk(file *os.File, chunkLen int) []byte {
	chunkData := readSomeBytes(file, int(chunkLen))
	return chunkData
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

func imageMaskFromRawData(data []byte, width int, height int) *image.Gray {
	img := image.NewGray(image.Rectangle{
		Min: image.Point{},
		Max: image.Point{X: width, Y: height},
	})
	img.Pix = data
	return img
}

func DoShit(shit ...interface{}) {

}

func ConvertMTXToPNG(file string) error {
	log.Info(file)

	// set up paths and file names for later
	fileDir, fileBase := filepath.Split(file)
	fileBaseNoExt := strings.Split(fileBase, ".")[0]

	// prepare PNG encoder
	pngEnc := png.Encoder{
		CompressionLevel: png.BestSpeed,
	}

	// open file
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}

	// parse file header
	fileHeader, err := readFileHeader(f)
	if err != nil {
		log.Fatal(err)
	}

	// do (optional?) size check
	if fi.Size() - 12 != int64(fileHeader.SecondImageOffset + fileHeader.SizeCheck) {
		log.Fatal("Size mismatch")
	}

	moreImagesAvailable := true
	imageIndex := 1
	for moreImagesAvailable {
		// create new file paths
		// newColorPath := filepath.Join(fileDir, fmt.Sprintf("%s_color%d.jpg", fileBaseNoExt, imageIndex))
		// newAlphaPath := filepath.Join(fileDir, fmt.Sprintf("%s_alpha%d.png", fileBaseNoExt, imageIndex))
		newOutFilePath := filepath.Join(fileDir, fmt.Sprintf("%s%d.png", fileBaseNoExt, imageIndex))

		log.Infof("Extracting image %d…\n", imageIndex)

		// get image header
		imgHeader, err := readImageHeader(f)
		if err != nil {
			log.Fatal(err)
		}
		log.Debug(imgHeader)

		// get color data
		chunkLength := int(binary.LittleEndian.Uint32(readSomeBytes(f, 4)))
		chunkData := readDataChunk(f, chunkLength)

		// and write it to file
		// err = ioutil.WriteFile(newColorPath, chunkData, 0644)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// create reader around the jpeg chunk
		chunkReader := bytes.NewReader(chunkData)

		// load image details without decoding the image
		colorImageConfig, colorImageFormat, err := image.DecodeConfig(chunkReader)
		if err != nil {
			log.Fatal(err)
		}

		log.Debugf("color%d decoded as %s\n", imageIndex, colorImageFormat)

		// if the image is bigger than the arbitrarily set limit, stop
		if colorImageConfig.Width > MAX_IMAGE_SIZE || colorImageConfig.Height > MAX_IMAGE_SIZE {
			log.Fatal("Image is larger than 4096 pixels on either the vertical or horizontal axis.")
		}

		// reset reader to the beginning
		chunkReader.Seek(0, io.SeekStart)

		// actually decode the image
		colorImage, colorImageFormat, err := image.Decode(bytes.NewReader(chunkData))

		pos, _ := f.Seek(0, io.SeekCurrent)
		log.Debugf("Position (after color%d): %d\n", imageIndex, pos)

		// get alpha data
		chunkLength = int(binary.LittleEndian.Uint32(readSomeBytes(f, 4)))
		chunkData = readDataChunk(f, chunkLength)

		pos, _ = f.Seek(0, io.SeekCurrent)
		log.Debugf("Position (after alpha%d): %d\n", imageIndex, pos)

		chunkDataDecompressed, err := decompressZlibData(chunkData)
		if err != nil {
			log.Fatal(err)
		}

		alphaImage := imageMaskFromRawData(chunkDataDecompressed, int(imgHeader.Width), int(imgHeader.Height))

		if colorImageConfig.Width != int(imgHeader.Width) || colorImageConfig.Height != int(imgHeader.Height) {
			log.Fatal("Size mismatch between color image and alpha mask")
		}

		bounds := colorImage.Bounds()
		rgba := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
		draw.Draw(rgba, bounds, colorImage, bounds.Min, draw.Src)

		for idx, alpha := range alphaImage.Pix {
			alphaIdx := idx * 4 + 3
			rgba.Pix[alphaIdx] = alpha
		}

		outImageFile, err := os.Create(newOutFilePath)
		if err != nil {
			log.Fatal(err)
		}

		err = pngEnc.Encode(outImageFile, rgba)
		if err != nil {
			log.Fatal(err)
		}

		if pos == fi.Size() {
			// EOF has been reached
			log.Info("All images extracted!")
			moreImagesAvailable = false
		}

		imageIndex++
	}

	return nil
}
