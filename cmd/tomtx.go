package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// tomtxCmd represents the tomtx command
var tomtxCmd = &cobra.Command{
	Use:   "tomtx [image files]",
	Short: "Convert images to MTX",

	Args: cobra.MinimumNArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("tomtx called")
	},
}

func init() {
	rootCmd.AddCommand(tomtxCmd)
}
