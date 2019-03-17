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
	WavFormat = Format(".wav")
	// Mp3Format represents .mp3 files.
	Mp3Format = Format(".mp3")
)

var (
	// WavBitDepths is the list of supported bit depths for wav format.
	WavBitDepths = map[signal.BitDepth]string{
		signal.BitDepth8:  "8 bit",
		signal.BitDepth16: "16 bits",
		signal.BitDepth24: "24 bits",
		signal.BitDepth32: "32 bits",
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
	Sink(Destination) pipe.Sink
}

// WavConfig is the configuration needed for wav output.
type WavConfig struct {
	signal.BitDepth
}

// Mp3Config is the configuration needed for mp3 output.
type Mp3Config struct{}

// Sink creates wav sink with provided config.
func (c WavConfig) Sink(d Destination) pipe.Sink {
	return wav.NewSink(d, c.BitDepth)
}

// Format returns wav format extension.
func (WavConfig) Format() Format {
	return WavFormat
}

// Sink creates mp3 sink with provided config.
func (c Mp3Config) Sink(d Destination) pipe.Sink {
	return mp3.NewSink(d, 192, 0)
}

// Format returns mp3 format extension.
func (Mp3Config) Format() Format {
	return Mp3Format
}

func (f Format) pump(s Source) (pipe.Pump, error) {
	switch f {
	case WavFormat:
		return wav.NewPump(s), nil
	case Mp3Format:
		return mp3.NewPump(s), nil
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
	sink := destinationConfig.Sink(d)

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
