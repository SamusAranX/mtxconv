package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"mtxconv/mtx"
)

// topngCmd represents the topng command
var topngCmd = &cobra.Command{
	Use:   "topng [PNG files to convert]",
	Short: "Convert MTX to PNG",

	Args: cobra.MinimumNArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "15:04:05",
		})

		if debugModeEnabled {
			log.SetLevel(log.DebugLevel)
			log.Debug("Debug flag set!")
		}

		for _, file := range args {
			if err := mtx.ConvertMTXToPNG(file, dryRunEnabled); err != nil {
				log.Error(err)
			}
			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(topngCmd)
}
