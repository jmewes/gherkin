package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var SourceParameter string
var TargetParameter string

var rootCmd = &cobra.Command{
	Use:   "gherkin",
	Short: "Command-line tool for processing feature files",
}

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
