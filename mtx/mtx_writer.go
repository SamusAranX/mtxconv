package mtx

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func placeholder(stuff ...interface{}) {}

func createMTXv0(inFile *os.File, inFileInfo fs.FileInfo, outFilePath string, dryRun bool) error {
	return errors.New("not implemented")
}

func createMTXv1(inFile *os.File, inFileInfo fs.FileInfo, outFilePath string, dryRun bool) error {
	return errors.New("not implemented")
}

func createMTXv2(inFile *os.File, outFilePath string, dryRun bool) error {
	mtxv2Header := []byte{2, 0, 0, 0, 0, 1}

	f, err := os.Open(outFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	inFileContents, err := ioutil.ReadAll(inFile)
	if err != nil {
		return err
	}

	f.Write(mtxv2Header)
	f.Write(inFileContents)

	return nil
}

func CreateMTXFile(file string, mtxTargetVersion int, dryRun bool) error {
	if mtxTargetVersion < -1 || mtxTargetVersion > 2 {
		return errors.New(fmt.Sprintf("an MTX target version of %d is unsupported. Supported values are: -1, 0, 1, and 2", mtxTargetVersion))
	}

	fileDir, fileBase := filepath.Split(file)
	fileNameSplit := strings.Split(fileBase, ".")
	_, fileExt := fileNameSplit[0], strings.ToLower(fileNameSplit[len(fileNameSplit)-1])
	newOutFilePath := filepath.Join(fileDir, fmt.Sprintf("%s.mtx", fileBase))

	// Do preflight checks here so they won't have to be repeated in the other functions
	if fileExt == ".mtx" {
		return errors.New("already an MTX file")
	} else if fileExt == ".jpeg" || fileExt == ".jpg" {
		if mtxTargetVersion == -1 {
			mtxTargetVersion = 0
		}
		if mtxTargetVersion == 2 {
			return errors.New("JPG files are only supported with MTX target versions 0 or 1")
		}
	} else if fileExt == ".png" {
		if mtxTargetVersion == -1 {
			mtxTargetVersion = 1
		}
		if mtxTargetVersion != 1 {
			return errors.New("PNG files are only supported with MTX target version 0 or 1")
		}
	} else if fileExt == ".pvr" {
		if mtxTargetVersion == -1 {
			mtxTargetVersion = 2
		}
		if mtxTargetVersion != 2 {
			return errors.New("PVR files are only supported with MTX target version 2")
		}
	} else {
		return errors.New("unsupported file format")
	}

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

	switch mtxTargetVersion {
	case 0:
		log.Debug("Format: MTXv0")
		if err := createMTXv0(f, fi, newOutFilePath, dryRun); err != nil {
			return err
		}
	case 1:
		log.Debug("Format: MTXv1")
		if err := createMTXv1(f, fi, newOutFilePath, dryRun); err != nil {
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

	placeholder("TODO: Implement this", newOutFilePath)

	return nil
}
