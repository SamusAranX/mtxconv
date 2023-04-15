package mtx

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/disintegration/imaging"
	log "github.com/sirupsen/logrus"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	jpegOpts *jpeg.Options = nil
)

func createMTXv0(inFile *os.File, outFilePath string, dryRun bool) error {
	img, err := imaging.Open(inFile.Name())
	if err != nil {
		return err
	}

	imgBuf := new(bytes.Buffer)
	err = jpeg.Encode(imgBuf, img, jpegOpts)
	if err != nil {
		return err
	}

	scaledImg := imaging.Resize(img, img.Bounds().Dx()/2, img.Bounds().Dy()/2, imaging.CatmullRom)
	scaledImgBuf := new(bytes.Buffer)
	err = jpeg.Encode(scaledImgBuf, scaledImg, jpegOpts)
	if err != nil {
		return err
	}

	fileHeader := HeaderV0V1{
		Magic:        0,
		LengthFirst:  uint32(len(scaledImgBuf.Bytes())),
		LengthSecond: uint32(len(imgBuf.Bytes())),
	}

	if dryRun {
		log.Debugf("Dry Run: skipping creation of %s", filepath.Base(outFilePath))
	} else {
		f, err := os.Create(outFilePath)
		if err != nil {
			return err
		}
		defer f.Close()

		binary.Write(f, binary.LittleEndian, fileHeader)
		f.Write(scaledImgBuf.Bytes())
		f.Write(imgBuf.Bytes())
	}

	return nil
}

func createMTXv1(inFile *os.File, outFilePath string, dryRun bool) error {
	rawImage, err := imaging.Open(inFile.Name())
	if err != nil {
		return err
	}

	// convert input image to NRGBA space and create a scaled-down copy
	img := imageToNRGBA(rawImage)
	scaledImg := imaging.Resize(img, img.Bounds().Dx()/2, img.Bounds().Dy()/2, imaging.CatmullRom)

	// compress the original image's alpha channel into a memory buffer using zlib
	originalAlphaCompressed, err := compressZlibData(getAlphaChannel(img))
	if err != nil {
		return err
	}

	// compress the scaled image's alpha channel into a memory buffer using zlib
	scaledAlphaCompressed, err := compressZlibData(getAlphaChannel(scaledImg))
	if err != nil {
		return err
	}

	// make both images' alpha channels fully opaque so the JPEG encoding step doesn't mess with transparent pixels
	makeAlphaChannelOpaque(img)
	makeAlphaChannelOpaque(scaledImg)

	// JPEG-encode original image into memory buffer
	originalBuf := new(bytes.Buffer)
	err = jpeg.Encode(originalBuf, img, jpegOpts)
	if err != nil {
		return err
	}

	// JPEG-encode scaled image into memory buffer
	scaledBuf := new(bytes.Buffer)
	err = jpeg.Encode(scaledBuf, scaledImg, jpegOpts)
	if err != nil {
		return err
	}

	// store buffer lengths for later
	// make sure these are uint32s because binary.Write will simply write zero bytes when these are ints
	originalBufLen := uint32(len(originalBuf.Bytes()))
	originalAlphaLen := uint32(len(originalAlphaCompressed))
	scaledBufLen := uint32(len(scaledBuf.Bytes()))
	scaledAlphaLen := uint32(len(scaledAlphaCompressed))

	fileHeader := HeaderV0V1{
		Magic: 1,
		// Length fields include block headers and chunk lengths
		LengthFirst:  BLOCK_HEADER_V1_SIZE + 8 + scaledBufLen + scaledAlphaLen,
		LengthSecond: BLOCK_HEADER_V1_SIZE + 8 + originalBufLen + originalAlphaLen,
	}

	blockHeader1 := BlockHeaderV1{
		Magic:  1,
		Width:  uint32(scaledImg.Bounds().Dx()),
		Height: uint32(scaledImg.Bounds().Dy()),
	}

	blockHeader2 := BlockHeaderV1{
		Magic:  1,
		Width:  uint32(img.Bounds().Dx()),
		Height: uint32(img.Bounds().Dy()),
	}

	if dryRun {
		log.Debugf("Dry Run: skipping creation of %s", filepath.Base(outFilePath))
	} else {
		f, err := os.Create(outFilePath)
		if err != nil {
			return err
		}
		defer f.Close()

		binary.Write(f, binary.LittleEndian, fileHeader)

		binary.Write(f, binary.LittleEndian, blockHeader1)
		binary.Write(f, binary.LittleEndian, scaledBufLen)
		f.Write(scaledBuf.Bytes())
		binary.Write(f, binary.LittleEndian, scaledAlphaLen)
		f.Write(scaledAlphaCompressed)

		binary.Write(f, binary.LittleEndian, blockHeader2)
		binary.Write(f, binary.LittleEndian, originalBufLen)
		f.Write(originalBuf.Bytes())
		binary.Write(f, binary.LittleEndian, originalAlphaLen)
		f.Write(originalAlphaCompressed)
	}

	return nil
}

func createMTXv2(inFile *os.File, outFilePath string, dryRun bool) error {
	fileHeader := HeaderV2{
		Magic:   2,
		Unknown: 256,
	}

	f, err := os.Create(outFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	inFileContents, err := io.ReadAll(inFile)
	if err != nil {
		return err
	}

	if dryRun {
		log.Debugf("Dry Run: skipping creation of %s", filepath.Base(outFilePath))
	} else {
		binary.Write(f, binary.LittleEndian, fileHeader)
		f.Write(inFileContents)
	}

	return nil
}

func CreateMTXFile(file string, mtxTargetVersion int, jpegQuality int, dryRun bool) error {
	if mtxTargetVersion < -1 || mtxTargetVersion > 2 {
		return errors.New(fmt.Sprintf("an MTX target version of %d is unsupported. Supported values are: -1, 0, 1, and 2", mtxTargetVersion))
	}

	fileDir, fileBase := filepath.Split(file)
	fileNameSplit := strings.Split(fileBase, ".")
	_, fileExt := fileNameSplit[0], strings.ToLower(fileNameSplit[len(fileNameSplit)-1])
	newOutFilePath := filepath.Join(fileDir, fmt.Sprintf("%s.mtx", fileBase))

	// Do preflight checks here so they won't have to be repeated in the other functions
	if fileExt == "mtx" {
		return errors.New("already an MTX file")
	} else if fileExt == "jpeg" || fileExt == "jpg" {
		if mtxTargetVersion == -1 {
			mtxTargetVersion = 0
		}
		if mtxTargetVersion == 2 {
			return errors.New("JPEG files are only supported with MTX target version 0 or 1")
		}
	} else if fileExt == "png" {
		if mtxTargetVersion == -1 {
			mtxTargetVersion = 1
		}

		if mtxTargetVersion == 2 {
			return errors.New("PNG files are only supported with MTX target version 0 or 1")
		}
	} else if fileExt == "pvr" {
		if mtxTargetVersion == -1 {
			mtxTargetVersion = 2
		}
		if mtxTargetVersion != 2 {
			return errors.New("PVR files are only supported with MTX target version 2")
		}
	} else {
		return errors.New("unsupported file format")
	}

	log.Debugf("Selected MTX format: %d", mtxTargetVersion)

	jpegOpts = &jpeg.Options{Quality: jpegQuality}

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
	} else if fi.Size() > MAX_INPUT_FILE_SIZE {
		return errors.New("file is larger than 1 GiB")
	}

	// by this point, only valid input files for any given MTX target versions should remain
	switch mtxTargetVersion {
	case 0:
		log.Debug("Format: MTXv0")
		if err := createMTXv0(f, newOutFilePath, dryRun); err != nil {
			return err
		}
	case 1:
		log.Debug("Format: MTXv1")
		if err := createMTXv1(f, newOutFilePath, dryRun); err != nil {
			return err
		}
	case 2:
		log.Debug("Format: MTXv2")
		if err := createMTXv2(f, newOutFilePath, dryRun); err != nil {
			return err
		}
	default:
		return errors.New("this isn't supposed to happen. please report this")
	}

	return nil
}
