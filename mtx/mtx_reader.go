package mtx

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	pngEnc = png.Encoder{
		CompressionLevel: png.BestSpeed,
	}
)

func HandleMTXv0(file *os.File, fileInfo fs.FileInfo, dryRun bool) error {
	// read MTX header
	fileHeader, err := readHeaderV0(file)
	if err != nil {
		return err
	}

	// set up paths and file names for later
	fileDir, fileBase := filepath.Split(file.Name())
	fileBaseNoExt := strings.Split(fileBase, ".")[0]

	// variables for use in the loop
	var chunkData []byte
	blockLengths := [2]int{
		int(fileHeader.FirstImageLength),
		int(fileHeader.SecondImageLength),
	}

	for i, length := range blockLengths {
		imageIndex := i + 1

		// create new file path
		newOutFilePath := filepath.Join(fileDir, fmt.Sprintf("%s%d.jpg", fileBaseNoExt, imageIndex))

		log.Infof("Extracting image %d…\n", imageIndex)

		if chunkData, err = readSomeBytes(file, length); err != nil {
			return err
		}

		if dryRun {
			log.Debugf("Dry Run: skipping creation of %s", filepath.Base(newOutFilePath))
		} else {
			outImageFile, err := os.Create(newOutFilePath)
			if err != nil {
				return err
			}

			_, err = outImageFile.Write(chunkData)
			if err != nil {
				outImageFile.Close()
				return err
			}

			outImageFile.Close()
		}
	}

	pos, _ := file.Seek(0, io.SeekCurrent)
	if pos < fileInfo.Size() {
		log.Warnf("There is additional data in the file after %d bytes!", pos)
	}

	log.Info("Done.")

	return nil
}

func HandleMTXv1(file *os.File, fileInfo fs.FileInfo, dryRun bool) error {
	// set up paths and file names for later
	fileDir, fileBase := filepath.Split(file.Name())
	fileBaseNoExt := strings.Split(fileBase, ".")[0]

	// get file size for later
	fileSize := fileInfo.Size()

	// read MTX header
	fileHeader, err := readHeaderV1(file)
	if err != nil {
		return err
	}

	// do (optional?) size check
	if fileSize-12 != int64(fileHeader.SecondBlockOffset+fileHeader.SizeCheck) {
		return errors.New("size verification failed")
	}

	// setting up variables that are gonna be reused throughout the loop
	var filePos int64
	var chunkLength int
	var chunkData []byte

	imageIndex := 1
	for filePos < fileSize {
		if imageIndex == 3 {
			log.Warn("There is additional data after the expected two image blocks.")
			log.Warn("Extraction will continue, but errors might occur.")
		}

		// create new file path
		newOutFilePath := filepath.Join(fileDir, fmt.Sprintf("%s%d.png", fileBaseNoExt, imageIndex))

		log.Infof("Extracting image %d…\n", imageIndex)

		// get image header
		blockHeader, err := readBlockHeaderV1(file)
		if err != nil {
			return err
		}

		// read the color data
		if b, err := readSomeBytes(file, 4); err != nil {
			return err
		} else {
			chunkLength = int(binary.LittleEndian.Uint32(b))
		}
		if chunkData, err = readSomeBytes(file, chunkLength); err != nil {
			return err
		}

		// create reader around the color data chunk
		chunkReader := bytes.NewReader(chunkData)

		// load image details without decoding the image
		colorImageConfig, colorImageFormat, err := image.DecodeConfig(chunkReader)
		if err != nil {
			return err
		}

		log.Debugf("color%d decoded as %s\n", imageIndex, colorImageFormat)

		// if the image is bigger than the arbitrarily set limit, stop
		if colorImageConfig.Width > MAX_IMAGE_SIZE || colorImageConfig.Height > MAX_IMAGE_SIZE {
			return errors.New("image is larger than 4096 pixels on either the vertical or horizontal axis")
		}

		// reset reader to the beginning and actually decode the image
		chunkReader.Seek(0, io.SeekStart)
		colorImage, colorImageFormat, err := image.Decode(chunkReader)

		filePos, _ = file.Seek(0, io.SeekCurrent)
		log.Debugf("Position (after color%d): %d\n", imageIndex, filePos)

		// get mask data
		if b, err := readSomeBytes(file, 4); err != nil {
			return err
		} else {
			chunkLength = int(binary.LittleEndian.Uint32(b))
		}
		if chunkData, err = readSomeBytes(file, chunkLength); err != nil {
			return err
		}

		filePos, _ = file.Seek(0, io.SeekCurrent)
		log.Debugf("Position (after alpha%d): %d\n", imageIndex, filePos)

		// decompress mask data and construct an image
		chunkDataDecompressed, err := decompressZlibData(chunkData)
		if err != nil {
			return err
		}

		maskImage := newGrayFromRawData(chunkDataDecompressed, int(blockHeader.Width), int(blockHeader.Height))
		if colorImageConfig.Width != int(blockHeader.Width) || colorImageConfig.Height != int(blockHeader.Height) {
			return errors.New("size mismatch between color image and alpha mask")
		}

		log.Infof("Image %d: %d×%d px", imageIndex, blockHeader.Width, blockHeader.Height)

		// convert color image to NRGBA and fill in the mask image's alpha values
		rgba := imageToNRGBA(colorImage)
		for idx, alpha := range maskImage.Pix {
			alphaIdx := idx*4 + 3
			rgba.Pix[alphaIdx] = alpha
		}

		if dryRun {
			log.Debugf("Dry Run: skipping creation of %s", filepath.Base(newOutFilePath))
		} else {
			outImageFile, err := os.Create(newOutFilePath)
			if err != nil {
				return err
			}

			err = pngEnc.Encode(outImageFile, rgba)
			if err != nil {
				outImageFile.Close()
				return err
			}

			outImageFile.Close()
		}

		imageIndex++
	}

	log.Info("Done.")

	return nil
}

func ConvertMTXToPNG(file string, dryRun bool) error {
	log.Info(file)

	// open file
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	// get file info and perform preliminary size check
	fi, err := f.Stat()
	if err != nil {
		return errors.New("couldn't get file info")
	} else if !fi.Mode().IsRegular() {
		return errors.New("is a directory")
	} else if fi.Size() < 24 {
		return errors.New("file is too small to be an MTX file")
	}

	// parse file header and run the appropriate converter
	fileVersionBytes, _ := readSomeBytes(f, 4)
	fileVersion := binary.LittleEndian.Uint32(fileVersionBytes)

	_, _ = f.Seek(0, io.SeekStart)

	switch fileVersion {
	case 0:
		log.Debug("Format: MTXv0")
		if err := HandleMTXv0(f, fi, dryRun); err != nil {
			return err
		}
	case 1:
		log.Debug("Format: MTXv1")
		if err := HandleMTXv1(f, fi, dryRun); err != nil {
			return err
		}
	case 2:
		return errors.New(fmt.Sprintf("Unsupported MTX version 0x%X", fileVersion))
	default:
		//return errors.New(fmt.Sprintf("Unsupported MTX version 0x%X", fileVersion))
	}

	return nil
}
