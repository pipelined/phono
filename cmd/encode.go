package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/pipelined/phono/encode"
	"github.com/pipelined/phono/input"
	"pipelined.dev/audio/fileformat"
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

func encodeCLI(ctx context.Context, paths []string, recursive bool, outDir string, bufferSize int, sink input.Sink, ext string) {
	if outDir != "" {
		if _, err := os.Stat(outDir); os.IsNotExist(err) {
			log.Printf("Out path doesn't exist: %v", err)
			return
		}
	}
	// build a map for easy-check
	mpaths := make(map[string]struct{})
	if !recursive {
		for _, p := range paths {
			mpaths[p] = struct{}{}
		}
	}

	command := "phono-encode"
	walkFn := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error during walk: %v\n", err)
		}
		if fi.IsDir() {
			// process subdirs
			if recursive {
				return nil
			}

			// if not recursive, skip all subdirs
			if _, ok := mpaths[path]; ok {
				return nil
			}
			return filepath.SkipDir
		}

		// try to parse format
		format, ok := fileformat.FormatByPath(path)
		if !ok {
			// file is not supported, skip
			return nil
		}

		// open file
		in, err := os.Open(path)
		if err != nil {
			log.Printf("Error opening file: %v\n", err)
			return nil
		}
		defer in.Close() // since we only read file, it's ok to close it with defer

		// create output filename
		var outFilename string
		if outDir != "" {
			outFilename = filepath.Join(outDir, outName("", command, ext))
		} else {
			outFilename = filepath.Join(filepath.Dir(path), outName("", command, ext))
		}

		out, err := os.Create(outFilename)
		if err != nil {
			log.Printf("Error creating output file: %v\n", err)
		}
		// error will be handled in the end of the flow
		defer out.Close()

		if err = encode.Run(ctx, bufferSize, format.Source(in), sink(out)); err != nil {
			return fmt.Errorf("Failed to execute pipe: %v", err)
		}
		return out.Close()
	}
	for _, path := range paths {
		err := filepath.Walk(path, walkFn)
		if err != nil {
			log.Print(err)
		}
	}
}

// outName generates an output file name with a next template:
// 	[prefix-]name-timestamp.ext
func outName(prefix, command, ext string) string {
	if prefix != "" {
		return fmt.Sprintf("%s-%s-%s%s", prefix, command, timestamp(), ext)
	}
	return fmt.Sprintf("%s-%s%s", command, timestamp(), ext)
}

func timestamp() string {
	return time.Now().Format("2006-01-02T150405.999")
}
