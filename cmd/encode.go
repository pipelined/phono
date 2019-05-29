package cmd

import (
	"github.com/spf13/cobra"
)

var (
	encodeCmd = &cobra.Command{
		Use:   "encode",
		Short: "Encode audio files",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
)

func init() {
	rootCmd.AddCommand(encodeCmd)
}
