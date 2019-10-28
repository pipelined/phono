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
		outPath    string
		recursive  bool
		bufferSize int
		bitDepth   int
	}{}
	encodeWavCmd = &cobra.Command{
		Use:                   "wav [flags] path...",
		DisableFlagsInUseLine: true,
		Short:                 "Encode audio files to wav format",
		Args:                  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// parse user input
			sink, err := file.WAVSink(encodeWav.bitDepth)
			if err != nil {
				log.Print(err)
				os.Exit(1)
			}
			// create channel for interruption and context for cancellation
			ctx, cancelFn := context.WithCancel(context.Background())
			interrupt := run(func() {
				// interrupt signal received, shut down
				cancelFn()
			})
			encode(ctx,
				args,
				encodeWav.recursive,
				encodeWav.outPath,
				encodeWav.bufferSize,
				sink,
				file.WAV.DefaultExtension(),
			)
			// block until interruption doesn't return
			<-interrupt
		},
	}
)

func init() {
	encodeCmd.AddCommand(encodeWavCmd)
	encodeWavCmd.Flags().StringVar(&encodeWav.outPath, "out", "", "output folder, the input folder is used if not specified")
	encodeWavCmd.Flags().IntVar(&encodeWav.bufferSize, "buffersize", 1024, "buffer size")
	encodeWavCmd.Flags().IntVar(&encodeWav.bitDepth, "bitdepth", 24, "bit depth")
	encodeWavCmd.Flags().BoolVar(&encodeWav.recursive, "recursive", false, "process paths recursive")
	encodeWavCmd.Flags().SortFlags = false
}
