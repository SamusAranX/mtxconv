package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"mtxconv/mtx"
)

var (
	mtxTargetVersion int
	jpegQuality      int
)

const (
	defaultJPEGQuality = 90 // estimated from extracted JPEG files
)

// bakeCmd represents the tomtx command
var bakeCmd = &cobra.Command{
	Use:   "bake [image files]",
	Short: "Convert images to MTX (not implemented yet)",

	Args: cobra.MinimumNArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		commandPreflight(debugModeEnabled)

		log.Debugf("bake called: %d", mtxTargetVersion)

		for _, file := range args {
			log.Info(file)
			if err := mtx.CreateMTXFile(file, mtxTargetVersion, jpegQuality, dryRunEnabled); err != nil {
				log.Error(err)
			}
			fmt.Println()
		}
	},
}

func init() {
	bakeCmd.Flags().IntVarP(&mtxTargetVersion, "mtx-version", "m", -1, "Target MTX version. Needs to be one of 0, 1, 2, or -1 to autoselect (Default -1)")
	bakeCmd.Flags().IntVarP(&jpegQuality, "jpeg-quality", "q", defaultJPEGQuality, fmt.Sprintf("JPEG quality (Default %d)", defaultJPEGQuality))
	rootCmd.AddCommand(bakeCmd)
}
