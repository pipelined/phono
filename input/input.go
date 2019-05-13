// Package input provides types to parse user input of pipe components.
package input

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pipelined/signal"

	"github.com/pipelined/mp3"
	"github.com/pipelined/pipe"
	"github.com/pipelined/wav"
)

type (
	// ConvertForm provides html form to the user. The form contains all information needed for conversion.
	ConvertForm interface {
		Data() []byte
		InputMaxSize(r *http.Request) (int64, error)
		ParsePump(r *http.Request) (pipe.Pump, io.Closer, error)
		ParseSink(r *http.Request) (BuildFunc, string, error)
	}

	wavFormat struct {
		DefaultExtension string
		Extensions       []string
		BitDepths        map[signal.BitDepth]struct{}
	}

	mp3Format struct {
		DefaultExtension string
		Extensions       []string
		MaxBitRate       int
		MinBitRate       int
		ChannelModes     map[mp3.ChannelMode]struct{}
	}

	// BuildFunc is used to inject WriteSeeker into Sink.
	BuildFunc func(io.WriteSeeker) pipe.Sink
)

var (
	// Wav provides logic required to process input of wav files.
	Wav = wavFormat{
		DefaultExtension: ".wav",
		Extensions:       []string{".wav", ".wave"},
		BitDepths: map[signal.BitDepth]struct{}{
			signal.BitDepth8:  {},
			signal.BitDepth16: {},
			signal.BitDepth24: {},
			signal.BitDepth32: {},
		},
	}

	// Mp3 provides logic required to process input of mp3 files.
	Mp3 = mp3Format{
		DefaultExtension: ".mp3",
		Extensions:       []string{".mp3"},
		MinBitRate:       8,
		MaxBitRate:       320,
		ChannelModes: map[mp3.ChannelMode]struct{}{
			mp3.JointStereo: {},
			mp3.Stereo:      {},
			mp3.Mono:        {},
		},
	}
)

// Build validates all parameters required to build wav sink. If valid, build closure is returned.
func (f wavFormat) Build(bitDepth signal.BitDepth) (BuildFunc, error) {
	if _, ok := f.BitDepths[bitDepth]; !ok {
		return nil, fmt.Errorf("Bit depth %v is not supported", bitDepth)
	}

	return func(ws io.WriteSeeker) pipe.Sink {
		return &wav.Sink{
			BitDepth:    bitDepth,
			WriteSeeker: ws,
		}
	}, nil
}

// Build validates all parameters required to build mp3 sink. If valid, build closure is returned.
func (f mp3Format) Build(bitRateMode mp3.BitRateMode, channelMode mp3.ChannelMode, useQuality bool, quality int) (BuildFunc, error) {
	if _, ok := f.ChannelModes[channelMode]; !ok {
		return nil, fmt.Errorf("Channel mode %v is not supported", channelMode)
	}

	switch brm := bitRateMode.(type) {
	case mp3.VBR:
		if brm.Quality < 0 || brm.Quality > 9 {
			return nil, fmt.Errorf("VBR quality %v is not supported", brm.Quality)
		}
	case mp3.CBR:
		if err := f.bitRate(brm.BitRate); err != nil {
			return nil, err
		}
	case mp3.ABR:
		if err := f.bitRate(brm.BitRate); err != nil {
			return nil, err
		}
	}

	if useQuality {
		if quality < 0 || quality > 9 {
			return nil, fmt.Errorf("MP3 quality %v is not supported", quality)
		}
	}

	return func(ws io.WriteSeeker) pipe.Sink {
		s := &mp3.Sink{
			BitRateMode: bitRateMode,
			ChannelMode: channelMode,
			Writer:      ws,
		}
		if useQuality {
			s.SetQuality(quality)
		}
		return s
	}, nil
}

// BitRate checks if provided bit rate is supported.
func (f mp3Format) bitRate(v int) error {
	if v > f.MaxBitRate || v < f.MinBitRate {
		return fmt.Errorf("Bit rate %v is not supported. Provide value between %d and %d", v, f.MinBitRate, f.MaxBitRate)
	}
	return nil
}
