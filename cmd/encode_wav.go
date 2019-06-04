package cmd

import (
	"context"
	"log"
	"os"

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
			// parse input
			buildFn, err := file.Wav.BuildSink(encodeWav.bitDepth)
			if err != nil {
				log.Print(err)
				os.Exit(1)
			}
			// create channel for interruption and context for cancellation
			interrupt := make(chan struct{})
			ctx, cancelFn := context.WithCancel(context.Background())
			go run(interrupt, func() {
				// interrupt signal received, shut down
				cancelFn()
			})
			encode(ctx, args, encodeWav.bufferSize, buildFn, file.Wav.DefaultExtension)
			// block until interruption doesn't return
			<-interrupt
		},
	}
)

func init() {
	encodeCmd.AddCommand(encodeWavCmd)
	encodeWavCmd.Flags().IntVar(&encodeWav.bitDepth, "bitdepth", 24, "bit depth")
	encodeWavCmd.Flags().IntVar(&encodeWav.bufferSize, "buffersize", 1024, "buffer size")
}
