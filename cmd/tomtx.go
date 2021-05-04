package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// tomtxCmd represents the tomtx command
var tomtxCmd = &cobra.Command{
	Use:   "tomtx",
	Short: "Convert PNG to MTX",

	Args: cobra.MinimumNArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("tomtx called")
	},
}

func init() {
	rootCmd.AddCommand(tomtxCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tomtxCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tomtxCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
