package cmd

import (
	"github.com/spf13/cobra"
)

var (
	workingDir string
	dryRun     bool
	backup     bool
	backupDir  string

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
	tfimpCmd.AddCommand(fromResourceCmd)
	tfimpCmd.PersistentFlags().StringVarP(&workingDir, "working-dir", "d", "./", "The working directory (default to current folder).")
	tfimpCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Run in dry-run mode.")
	tfimpCmd.PersistentFlags().BoolVar(&backup, "backup", false, "Whether to backup the state file before running the import commands.")
	tfimpCmd.PersistentFlags().StringVar(&backupDir, "backup-dir", ".", "Where to store the state backup file.")
}
