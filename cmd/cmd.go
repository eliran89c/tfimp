package cmd

import (
	"github.com/spf13/cobra"
)

var (
	workingDir string
	noDryRun   bool

	tfimpCmd = &cobra.Command{
		Use:           "tfimp",
		Short:         "tfimp - command-line tool to easily import bulk resources to terraform stacks",
		Long:          ``,
		Version:       "0.1.0",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func Execute() error {
	return tfimpCmd.Execute()
}

func init() {
	tfimpCmd.AddCommand(fromConfigFileCmd)
	tfimpCmd.PersistentFlags().StringVarP(&workingDir, "working-dir", "d", "./", "The working directory.")
	tfimpCmd.PersistentFlags().BoolVar(&noDryRun, "no-dry-run", false, "Disable dry-run mode.")
}
