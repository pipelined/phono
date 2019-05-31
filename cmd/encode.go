package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pipelined/phono/input"
	"github.com/pipelined/pipe"
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

func encodeFiles(bufferSize int, buildFn input.BuildFunc, ext string) filepath.WalkFunc {
	return func(path string, file os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error during walk: %v\n", err)
		}
		if file.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			log.Printf("Error opening file: %v\n", err)
			return nil
		}
		pump, err := input.FilePump(path, f)
		if err != nil {
			log.Printf("Error creating a pump: %v\n", err)
			return nil
		}
		dir, name := filepath.Split(path)
		result, err := os.Create(outFileName(dir, name, ext))
		if err != nil {
			log.Printf("Error creating output file: %v\n", err)
		}
		// build encode pipe
		encode, err := pipe.New(bufferSize,
			pipe.WithPump(pump),
			pipe.WithSinks(buildFn(result)),
		)
		if err != nil {
			return fmt.Errorf("Failed to build pipe: %v", err)
		}

		// run conversion
		err = pipe.Wait(encode.Run())
		if err != nil {
			return fmt.Errorf("Failed to execute pipe: %v", err)
		}
		return nil
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
