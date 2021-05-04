package cmd

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mtxconv",
	Short: "A converter to and from Mediocre's MTX format",
}

var (
	debugModeEnabled bool
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize()

	rootCmd.PersistentFlags().BoolVarP(&debugModeEnabled,"debug", "", false, "debug")
}