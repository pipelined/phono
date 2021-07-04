package cmd

import (
	"context"
	"log"
	"os"

	"github.com/spf13/cobra"
	"pipelined.dev/audio/fileformat"

	"pipelined.dev/phono/userinput"
)

var (
	encodeMp3 = struct {
		outPath     string
		recursive   bool
		bufferSize  int
		channelMode int
		bitRateMode string
		bitRate     int
		quality     int
	}{}
	encodeMp3Cmd = &cobra.Command{
		Use:                   "mp3 [flags] path...",
		DisableFlagsInUseLine: true,
		Short:                 "Encode audio files to mp3 format",
		Args:                  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			useQuality := false
			if cmd.Flags().Changed("quality") {
				useQuality = true
			}
			sink, err := userinput.MP3.Sink(
				encodeMp3.bitRateMode,
				encodeMp3.bitRate,
				encodeMp3.channelMode,
				useQuality,
				encodeMp3.quality,
			)
			if err != nil {
				log.Print(err)
				os.Exit(1)
			}
			// create channel for interruption and context for cancellation
			ctx, cancelFn := context.WithCancel(context.Background())
			// interrupt signal received, shut down
			interrupted := onInterrupt(func() { cancelFn() })
			encodeCLI(ctx,
				args,
				encodeMp3.recursive,
				encodeMp3.outPath,
				encodeMp3.bufferSize,
				sink,
				fileformat.MP3().DefaultExtension(),
			)
			<-interrupted
		},
	}
)

func init() {
	encodeCmd.AddCommand(encodeMp3Cmd)
	encodeMp3Cmd.Flags().StringVar(&encodeMp3.outPath, "out", "", "output folder, the userinput folder is used if not specified")
	encodeMp3Cmd.Flags().IntVar(&encodeMp3.bufferSize, "buffersize", 1024, "buffer size")
	encodeMp3Cmd.Flags().IntVar(&encodeMp3.channelMode, "channelmode", 2, "channel mode:\n0 - mono\n1 - stereo\n2 - joint stereo")
	encodeMp3Cmd.Flags().StringVar(&encodeMp3.bitRateMode, "bitratemode", "vbr", "bit rate mode:\ncbr - constant bit rate\nabr - average bit rate\nvbr - variable bit rate")
	encodeMp3Cmd.Flags().IntVar(&encodeMp3.bitRate, "bitrate", 4, "bit rate:\n[8..320] for cbr and abr\n[0..9] for vbr")
	encodeMp3Cmd.Flags().IntVar(&encodeMp3.quality, "quality", 5, "quality [0..9]")
	encodeMp3Cmd.Flags().BoolVar(&encodeMp3.recursive, "recursive", false, "process paths recursive")
	encodeMp3Cmd.Flags().SortFlags = false
}
