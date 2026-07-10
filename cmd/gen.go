package cmd

import (
	"fmt"
	"os"

	"github.com/experimental-software/gherkin/core"
	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use:   "gen [flags]",
	Short: "Generate source code from feature files",

	Run: func(cmd *cobra.Command, args []string) {
		doc, err := core.ParseFeatureFile(SourceParameter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(core.FeatureTitle(doc))
	},
}

func init() {
	genCmd.Flags().StringVarP(&SourceParameter, "source", "s", "", "Path to source file")
	_ = genCmd.MarkFlagRequired("source")

	rootCmd.AddCommand(genCmd)
}
