package constants

import (
	"log"
	"time"
)

var (
	GitBranch      = "n/a"
	GitCommit      = "n/a"
	GitCommitShort = "n/a"
	GitVersion     = "n/a"
	GitTag         = "v0.0.0" // provide default value that can be parsed
	BuildTime      = "n/a"
)

const (
	BuildTimeFormat = "2006-01-02T15:04:05 MST"
)

func ParseBuildTime() *time.Time {
	parsed, err := time.Parse(BuildTimeFormat, BuildTime)
	if err != nil {
		log.Printf("ParseBuildTime: %v", err.Error())
		return nil
	}

	return &parsed
}

func ParseBuildTimeLocal() *time.Time {
	parsed, err := time.Parse(BuildTimeFormat, BuildTime)
	if err != nil {
		log.Printf("ParseBuildTimeLocal: %v", err.Error())
		return nil
	}

	localLocation, err := time.LoadLocation("Local")
	if err != nil {
		log.Printf("ParseBuildTimeLocal: %v", err.Error())
		return nil
	}

	localTime := parsed.In(localLocation)
	return &localTime
}
