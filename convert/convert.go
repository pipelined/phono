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

// Destination is the output of convertation.
type Destination interface {
	io.Writer
	io.Seeker
}

// Format is a file extension.
type Format string

// OutputConfig is an interface that defines how Sink is created out of configuration.
type OutputConfig interface {
	Format() Format
	sink(Destination) (pipe.Sink, error)
}

// WavConfig is the configuration needed for wav output.
type WavConfig struct {
	signal.BitDepth
}

// Mp3Config is the configuration needed for mp3 output.
type Mp3Config struct {
	mp3.BitRateMode
	mp3.ChannelMode
	BitRate int
	mp3.VBRQuality
	UseQuality bool
	mp3.Quality
}

// Sink creates wav sink with provided config.
// Returns error if provided configuration is not supported.
func (c WavConfig) sink(d Destination) (pipe.Sink, error) {
	// check if bit depth is supported
	if _, ok := Supported.WavBitDepths[c.BitDepth]; !ok {
		return nil, fmt.Errorf("Bit depth %v is not supported", c.BitDepth)
	}

	return &wav.Sink{
		WriteSeeker: d,
		BitDepth:    c.BitDepth,
	}, nil
}

// Format returns wav format extension.
func (WavConfig) Format() Format {
	return WavFormat
}

// Sink creates mp3 sink with provided config.
func (c Mp3Config) sink(d Destination) (pipe.Sink, error) {
	// check if bit rate mode is supported
	if _, ok := Supported.Mp3BitRateModes[c.BitRateMode]; !ok {
		return nil, fmt.Errorf("Bit rate mode %v is not supported", c.BitRateMode)
	}

	// check if channel mode is supported
	if _, ok := Supported.Mp3ChannelModes[c.ChannelMode]; !ok {
		return nil, fmt.Errorf("Channel mode %v is not supported", c.ChannelMode)
	}

	// check if quality is supported
	if c.UseQuality {
		if _, ok := Supported.Mp3Qualities[c.Quality]; !ok {
			return nil, fmt.Errorf("Quality %v is not supported", c.Quality)
		}
	}

	if c.BitRateMode == mp3.VBR {
		// validate VBR quality
		if _, ok := Supported.Mp3VBRQualities[c.VBRQuality]; !ok {
			return nil, fmt.Errorf("VBR quality %v is not supported", c.VBRQuality)
		}
	} else {
		// validate bit rate for ABR and CBR
		if c.BitRate > mp3.MaxBitRate || c.BitRate < mp3.MinBitRate {
			return nil, fmt.Errorf("Bit rate %v is not supported. Provide value between %d and %d", c.BitRate, mp3.MinBitRate, mp3.MaxBitRate)
		}
	}

	var s mp3.Sink
	switch c.BitRateMode {
	case mp3.CBR:
		s = &mp3.CBRSink{
			Writer:      d,
			ChannelMode: c.ChannelMode,
			BitRate:     c.BitRate,
		}
	case mp3.ABR:
		s = &mp3.ABRSink{
			Writer:      d,
			ChannelMode: c.ChannelMode,
			BitRate:     c.BitRate,
		}
	case mp3.VBR:
		s = &mp3.VBRSink{
			Writer:      d,
			ChannelMode: c.ChannelMode,
			VBRQuality:  c.VBRQuality,
		}
	}
	if c.UseQuality {
		s.SetQuality(c.Quality)
	}
	return s, nil
}

// Format returns mp3 format extension.
func (Mp3Config) Format() Format {
	return Mp3Format
}

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
func Convert(s Source, d Destination, sourceFormat Format, destinationConfig OutputConfig) error {
	// create pump for input format
	pump, err := sourceFormat.pump(s)
	if err != nil {
		return fmt.Errorf("Unsupported input format: %s", sourceFormat)
	}
	// create sink for output format
	sink, err := destinationConfig.sink(d)
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
