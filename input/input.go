// Package input provides types to parse user input for pipe components.
package input

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/pipelined/signal"

	"github.com/pipelined/mp3"
	"github.com/pipelined/pipe"
	"github.com/pipelined/wav"
)

type (
	// ConvertForm provides html form to the user. The form contains all information needed for conversion.
	ConvertForm interface {
		Data() []byte
		InputMaxSize(url string) (int64, error)
		FileKey() string
		ParseSink(data url.Values) (BuildFunc, string, error)
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
		VBR              string
		CBR              string
		ABR              string
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
		VBR: "VBR",
		ABR: "ABR",
		CBR: "CBR",
	}
)

// FilePump returns pump for provided file source. Type of the pump is determined by file extension.
func FilePump(fileName string, rs io.ReadSeeker) (pipe.Pump, error) {
	switch {
	case HasExtension(fileName, Wav.Extensions):
		return &wav.Pump{ReadSeeker: rs}, nil
	case HasExtension(fileName, Mp3.Extensions):
		return &mp3.Pump{Reader: rs}, nil
	default:
		return nil, fmt.Errorf("File has unsupported extension: %v", fileName)
	}
}

// Build validates all parameters required to build wav sink. If valid, build closure is returned.
func (f wavFormat) Build(bitDepth int) (BuildFunc, error) {
	bd := signal.BitDepth(bitDepth)
	if _, ok := f.BitDepths[bd]; !ok {
		return nil, fmt.Errorf("Bit depth %v is not supported", bitDepth)
	}

	return func(ws io.WriteSeeker) pipe.Sink {
		return &wav.Sink{
			BitDepth:    bd,
			WriteSeeker: ws,
		}
	}, nil
}

// Build validates all parameters required to build mp3 sink. If valid, build closure is returned.
func (f mp3Format) Build(bitRateMode string, bitRate, channelMode int, useQuality bool, quality int) (BuildFunc, error) {
	cm := mp3.ChannelMode(channelMode)
	if _, ok := f.ChannelModes[cm]; !ok {
		return nil, fmt.Errorf("Channel mode %v is not supported", cm)
	}

	var brm mp3.BitRateMode
	switch bitRateMode {
	case f.VBR:
		if bitRate < 0 || bitRate > 9 {
			return nil, fmt.Errorf("VBR quality %v is not supported", bitRate)
		}
		brm = mp3.VBR(bitRate)
	case f.CBR:
		if err := f.bitRate(bitRate); err != nil {
			return nil, err
		}
		brm = mp3.CBR(bitRate)
	case f.ABR:
		if err := f.bitRate(bitRate); err != nil {
			return nil, err
		}
		brm = mp3.ABR(bitRate)
	default:
		return nil, fmt.Errorf("VBR mode %v is not supported", bitRateMode)
	}

	if useQuality {
		if quality < 0 || quality > 9 {
			return nil, fmt.Errorf("MP3 quality %v is not supported", quality)
		}
	}

	return func(ws io.WriteSeeker) pipe.Sink {
		s := &mp3.Sink{
			BitRateMode: brm,
			ChannelMode: cm,
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

// func (f mp3Format) BitRateMode(bitRateMode string, value int) mp3.BitRateMode {}

// HasExtension validates if filename has one of passed extensions.
// Filename is lower-cased before comparison.
func HasExtension(fileName string, exts []string) bool {
	name := strings.ToLower(fileName)
	for _, ext := range exts {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}
