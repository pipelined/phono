package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pipelined/phono/file"
	"github.com/pipelined/phono/pipes"

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

func encode(ctx context.Context, paths []string, bufferSize int, buildFn file.BuildSinkFunc, ext string) {
	walkFn := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error during walk: %v\n", err)
		}
		if fi.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			log.Printf("Error opening file: %v\n", err)
			return nil
		}
		pump, err := file.Pump(path, f)
		if err != nil {
			log.Printf("Cannot create a pump: %v\n", err)
			return nil
		}
		dir, name := filepath.Split(path)
		result, err := os.Create(outFileName(dir, name, ext))
		if err != nil {
			log.Printf("Error creating output file: %v\n", err)
		}

		if err = pipes.Encode(ctx, bufferSize, pump, buildFn(result)); err != nil {
			return fmt.Errorf("Failed to execute pipe: %v", err)
		}
		return nil
	}
	for _, path := range paths {
		err := filepath.Walk(path, walkFn)
		if err != nil {
			log.Print(err)
		}
	}
}

func outFileName(dir, name, ext string) string {
	n := time.Now()
	if dir == "" {
		// return ""
		return fmt.Sprintf("%s_%02d%02d%02d_%-3d%s", name, n.Hour(), n.Minute(), n.Second(), n.Nanosecond()/int(time.Millisecond), ext)
	}
	return fmt.Sprintf("%s%s_%02d%02d%02d_%-3d%s", dir, name, n.Hour(), n.Minute(), n.Second(), n.Nanosecond()/int(time.Millisecond), ext)

}
