package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/pipelined/phono/file"
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
			buildFn, err := file.Wav.BuildSink(encodeWav.bitDepth)
			if err != nil {
				log.Print(err)
				os.Exit(1)
			}
			walkFn := encode(encodeWav.bufferSize, buildFn, file.Wav.DefaultExtension)
			for _, path := range args {
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
