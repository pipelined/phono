package convert

import (
	"fmt"
	"io"

	"github.com/pipelined/mp3"
	"github.com/pipelined/pipe"
	"github.com/pipelined/signal"
	"github.com/pipelined/wav"
)

const (
	// WavFormat represents .wav files.
	WavFormat Format = "wav"
	// Mp3Format represents .mp3 files.
	Mp3Format Format = "mp3"
)

var (
	// supported bit depths for wav format.
	wavBitDepths = map[signal.BitDepth]string{
		signal.BitDepth8:  "8 bit",
		signal.BitDepth16: "16 bits",
		signal.BitDepth24: "24 bits",
		signal.BitDepth32: "32 bits",
	}

	// Supported bit rate modes for mp3 format.
	mp3BitRateModes = map[mp3.BitRateMode]string{
		mp3.CBR: mp3.CBR.String(),
		mp3.VBR: mp3.VBR.String(),
		mp3.ABR: mp3.ABR.String(),
	}

	// Supported channel modes for mp3 format.
	mp3ChannelModes = map[mp3.ChannelMode]string{
		mp3.JointStereo: "joint stereo",
		mp3.Stereo:      "stereo",
		mp3.Mono:        "mono",
	}

	// Supported VBR quality values for mp3 format.
	mp3VBRQualities = map[mp3.VBRQuality]string{
		mp3.VBR0: "0",
		mp3.VBR1: "1",
		mp3.VBR2: "2",
		mp3.VBR3: "3",
		mp3.VBR4: "4",
		mp3.VBR5: "5",
		mp3.VBR6: "6",
		mp3.VBR7: "7",
		mp3.VBR8: "8",
		mp3.VBR9: "9",
	}

	// mp3Qualities is the list of supported VBR quality values for mp3 format.
	mp3Qualities = map[mp3.Quality]string{
		mp3.Q0: "0",
		mp3.Q1: "1",
		mp3.Q2: "2",
		mp3.Q3: "3",
		mp3.Q4: "4",
		mp3.Q5: "5",
		mp3.Q6: "6",
		mp3.Q7: "7",
		mp3.Q8: "8",
		mp3.Q9: "9",
	}

	// Supported values for convert configuration.
	Supported = struct {
		WavBitDepths    map[signal.BitDepth]string
		Mp3BitRateModes map[mp3.BitRateMode]string
		Mp3ChannelModes map[mp3.ChannelMode]string
		Mp3VBRQualities map[mp3.VBRQuality]string
		Mp3Qualities    map[mp3.Quality]string
	}{
		WavBitDepths:    wavBitDepths,
		Mp3BitRateModes: mp3BitRateModes,
		Mp3ChannelModes: mp3ChannelModes,
		Mp3Qualities:    mp3Qualities,
		Mp3VBRQualities: mp3VBRQualities,
	}
)

// Source is the input for convertation.
type Source interface {
	io.Reader
	io.Seeker
	io.Closer
}

// Format is a file extension.
type Format string

func (f Format) pump(s Source) (pipe.Pump, error) {
	switch f {
	case WavFormat:
		return &wav.Pump{ReadSeeker: s}, nil
	case Mp3Format:
		return &mp3.Pump{Reader: s}, nil
	default:
		return nil, fmt.Errorf("Unsupported format: %v", f)
	}
}

// Convert provided source of sourceFormat into destination using destinationConfig.
func Convert(s Source, sourceFormat Format, destinationConfig OutputConfig) error {
	// create pump for input format
	pump, err := sourceFormat.pump(s)
	if err != nil {
		return fmt.Errorf("Unsupported input format: %s", sourceFormat)
	}
	// create sink for output format
	sink, err := destinationConfig.sink()
	if err != nil {
		return fmt.Errorf("Provided configuration is not supported")
	}

	// build convert pipe
	convert, err := pipe.New(1024, pipe.WithPump(pump), pipe.WithSinks(sink))
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
