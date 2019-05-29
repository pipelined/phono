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
	encodeWav = struct {
		bufferSize int
		bitDepth   int
	}{}
	encodeWavCmd = &cobra.Command{
		Use:   "wav",
		Short: "Encode audio files to wav format",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			buildFn, err := input.Wav.Build(encodeWav.bitDepth)
			if err != nil {
				log.Print(err)
				os.Exit(1)
			}
			walkFn := encodeToWav(encodeWav.bufferSize, buildFn, input.Wav.DefaultExtension)
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
	encodeCmd.AddCommand(encodeWavCmd)
	encodeWavCmd.Flags().IntVar(&encodeWav.bitDepth, "bitdepth", 24, "bit depth")
	encodeWavCmd.Flags().IntVar(&encodeWav.bufferSize, "buffersize", 1024, "buffer size")
}

func encodeToWav(bufferSize int, buildFn input.BuildFunc, ext string) filepath.WalkFunc {
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
