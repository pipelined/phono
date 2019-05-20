package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "phono",
	Short: "DSP pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

// Execute the root comand.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
