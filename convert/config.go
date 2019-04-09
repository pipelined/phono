package convert

import (
	"fmt"
	"io"

	"github.com/pipelined/mp3"
	"github.com/pipelined/pipe"
	"github.com/pipelined/signal"
	"github.com/pipelined/wav"
)

// OutputConfig is an interface that defines how Sink is created out of configuration.
type OutputConfig interface {
	Format() Format
	sink() (pipe.Sink, error)
}

// WavConfig is the configuration needed for wav output.
type WavConfig struct {
	io.WriteSeeker
	signal.BitDepth
}

// Mp3Config is the configuration needed for mp3 output.
type Mp3Config struct {
	io.Writer
	mp3.BitRateMode
	mp3.ChannelMode
	BitRate int
	mp3.VBRQuality
	UseQuality bool
	mp3.Quality
}

// Sink creates wav sink with provided config.
// Returns error if provided configuration is not supported.
func (c *WavConfig) sink() (pipe.Sink, error) {
	// check if bit depth is supported
	if _, ok := Supported.WavBitDepths[c.BitDepth]; !ok {
		return nil, fmt.Errorf("Bit depth %v is not supported", c.BitDepth)
	}

	return &wav.Sink{
		WriteSeeker: c.WriteSeeker,
		BitDepth:    c.BitDepth,
	}, nil
}

// Format returns wav format extension.
func (*WavConfig) Format() Format {
	return WavFormat
}

// Sink creates mp3 sink with provided config.
func (c *Mp3Config) sink() (pipe.Sink, error) {
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
			Writer:      c.Writer,
			ChannelMode: c.ChannelMode,
			BitRate:     c.BitRate,
		}
	case mp3.ABR:
		s = &mp3.ABRSink{
			Writer:      c.Writer,
			ChannelMode: c.ChannelMode,
			BitRate:     c.BitRate,
		}
	case mp3.VBR:
		s = &mp3.VBRSink{
			Writer:      c.Writer,
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
func (*Mp3Config) Format() Format {
	return Mp3Format
}
