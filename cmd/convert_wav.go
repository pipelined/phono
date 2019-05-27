package cmd

import (
	"fmt"
	"github.com/pipelined/pipe"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/pipelined/phono/input"
)

var (
	convertWav = struct {
		bufferSize int
		bitDepth   int
	}{}
	convertWavCmd = &cobra.Command{
		Use:   "wav",
		Short: "Convert audio files to wav format",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			buildFn, err := input.Wav.Build(convertWav.bitDepth)
			if err != nil {
				log.Print(err)
				os.Exit(1)
			}
			walkFn := convertToWav(convertWav.bufferSize, buildFn, input.Wav.DefaultExtension)
			for _, path := range args {
				fmt.Printf("Dir: %v", path)
				err := filepath.Walk(path, walkFn)
				if err != nil {
					log.Print(err)
				}
			}
		},
	}
)

func init() {
	convertCmd.AddCommand(convertWavCmd)
	convertWavCmd.Flags().IntVar(&convertWav.bitDepth, "bitdepth", 24, "bit depth")
	convertWavCmd.Flags().IntVar(&convertWav.bufferSize, "buffersize", 1024, "buffer size")
}

func convertToWav(bufferSize int, buildFn input.BuildFunc, ext string) filepath.WalkFunc {
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
		// build convert pipe
		convert, err := pipe.New(bufferSize,
			pipe.WithPump(pump),
			pipe.WithSinks(buildFn(result)),
		)
		if err != nil {
			return fmt.Errorf("Failed to build pipe: %v", err)
		}

		// run conversion
		err = pipe.Wait(convert.Run())
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
