package cmd

import (
	"github.com/spf13/cobra"
)

var (
	convertCmd = &cobra.Command{
		Use:   "convert",
		Short: "Convert audio files",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
)

func init() {
	rootCmd.AddCommand(convertCmd)
}
