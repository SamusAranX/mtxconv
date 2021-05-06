package cmd

import log "github.com/sirupsen/logrus"

func commandPreflight(debugMode bool) {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05",
	})

	if debugMode {
		log.SetLevel(log.DebugLevel)
		log.Debug("Debug flag set!")
	}
}
