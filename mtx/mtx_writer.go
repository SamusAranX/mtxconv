package mtx

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

func placeholder(stuff ...interface{}) {}

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
			return errors.New("JPG files are not supported with MTX target version 2")
		}
	} else if fileExt == ".png" {
		if mtxTargetVersion == -1 {
			mtxTargetVersion = 1
		}
		if mtxTargetVersion != 1 {
			return errors.New("PNG files are only supported with MTX target version 1")
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

	placeholder("TODO: Implement this", newOutFilePath)

	return nil
}
