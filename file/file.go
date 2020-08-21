package file

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"pipelined.dev/audio/flac"
	"pipelined.dev/audio/mp3"
	"pipelined.dev/audio/wav"
	"pipelined.dev/pipe"
	"pipelined.dev/signal"
)

type (
	// Format represents audio file format.
	Format interface {
		Pump(io.ReadSeeker) pipe.Pump
		DefaultExtension() string
		Extensions() map[string]struct{}
	}

	wavFormat struct {
		extensions
		BitDepths map[signal.BitDepth]struct{}
	}

	mp3Format struct {
		extensions
		ChannelModes map[mp3.ChannelMode]struct{}
		VBR          string
		CBR          string
		ABR          string
		MaxBitRate   int
		MinBitRate   int
		MinQuality   int
		MaxQuality   int
		MinVBR       int
		MaxVBR       int
	}

	flacFormat struct {
		extensions
	}

	// generic struct to provide format with extensions.
	extensions struct {
		standard string
		all      map[string]struct{}
	}

	// Sink is used to inject WriteSeeker into Sink.
	Sink func(io.WriteSeeker) pipe.Sink
)

var (
	// WAV provides structures required to handle wav files.
	WAV = wavFormat{
		extensions: extensions{
			standard: ".wav",
			all: map[string]struct{}{
				".wav":  {},
				".wave": {},
			},
		},
		BitDepths: map[signal.BitDepth]struct{}{
			signal.BitDepth8:  {},
			signal.BitDepth16: {},
			signal.BitDepth24: {},
			signal.BitDepth32: {},
		},
	}

	// MP3 provides structures required to handle mp3 files.
	MP3 = mp3Format{
		extensions: extensions{
			standard: ".mp3",
			all: map[string]struct{}{
				".mp3": {},
			},
		},
		ChannelModes: map[mp3.ChannelMode]struct{}{
			mp3.JointStereo: {},
			mp3.Stereo:      {},
			mp3.Mono:        {},
		},
		VBR:        "VBR",
		ABR:        "ABR",
		CBR:        "CBR",
		MinBitRate: 8,
		MaxBitRate: 320,
		MinQuality: 0,
		MaxQuality: 9,
		MinVBR:     0,
		MaxVBR:     9,
	}

	// FLAC provides structures required to handle flac files.
	FLAC = flacFormat{
		extensions: extensions{
			standard: ".flac",
			all: map[string]struct{}{
				".flac": {},
			},
		},
	}
)

// ParseFormat determines file format by file extension.
func ParseFormat(fileName string) (Format, error) {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch {
	case hasExtension(ext, WAV.extensions.all):
		return WAV, nil
	case hasExtension(ext, MP3.extensions.all):
		return MP3, nil
	case hasExtension(ext, FLAC.extensions.all):
		return FLAC, nil
	default:
		return nil, fmt.Errorf("File has unsupported extension: %v", fileName)
	}
}

// hasExtension validates if filename has one of passed extensions.
// Filename is lower-cased before comparison.
func hasExtension(ext string, exts map[string]struct{}) bool {
	_, ok := exts[ext]
	return ok
}

// WAVSink validates all parameters required to build wav sink. If valid, build closure is returned.
// Closure allows to postpone io opertaions and do them only after all sink parameters are validated.
func WAVSink(bitDepth int) (Sink, error) {
	bd := signal.BitDepth(bitDepth)
	if _, ok := WAV.BitDepths[bd]; !ok {
		return nil, fmt.Errorf("Bit depth %v is not supported", bitDepth)
	}

	return func(ws io.WriteSeeker) pipe.Sink {
		return &wav.Sink{
			BitDepth:    bd,
			WriteSeeker: ws,
		}
	}, nil
}

func (wavFormat) Pump(rs io.ReadSeeker) pipe.Pump {
	return &wav.Pump{ReadSeeker: rs}
}

// MP3Sink validates all parameters required to build mp3 sink. If valid, Sink closure is returned.
// Closure allows to postpone io opertaions and do them only after all sink parameters are validated.
func MP3Sink(bitRateMode string, bitRate, channelMode int, useQuality bool, quality int) (Sink, error) {
	cm := mp3.ChannelMode(channelMode)
	if _, ok := MP3.ChannelModes[cm]; !ok {
		return nil, fmt.Errorf("Channel mode %v is not supported", cm)
	}

	var brm mp3.BitRateMode
	switch strings.ToUpper(bitRateMode) {
	case MP3.VBR:
		if bitRate < MP3.MinVBR || bitRate > MP3.MaxVBR {
			return nil, fmt.Errorf("VBR quality %v is not supported", bitRate)
		}
		brm = mp3.VBR(bitRate)
	case MP3.CBR:
		if err := MP3.bitRate(bitRate); err != nil {
			return nil, err
		}
		brm = mp3.CBR(bitRate)
	case MP3.ABR:
		if err := MP3.bitRate(bitRate); err != nil {
			return nil, err
		}
		brm = mp3.ABR(bitRate)
	default:
		return nil, fmt.Errorf("Bit rate mode %v is not supported", bitRateMode)
	}

	if useQuality {
		if quality < MP3.MinQuality || quality > MP3.MaxQuality {
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

func (mp3Format) Pump(rs io.ReadSeeker) pipe.Pump {
	return &mp3.Pump{Reader: rs}
}

// BitRate checks if provided bit rate is supported.
func (f mp3Format) bitRate(v int) error {
	if v > f.MaxBitRate || v < f.MinBitRate {
		return fmt.Errorf("Bit rate %v is not supported. Provide value between %d and %d", v, f.MinBitRate, f.MaxBitRate)
	}
	return nil
}

func (flacFormat) Pump(rs io.ReadSeeker) pipe.Pump {
	return &flac.Pump{Reader: rs}
}

func (e extensions) DefaultExtension() string {
	return e.standard
}

func (e extensions) Extensions() map[string]struct{} {
	exts := make(map[string]struct{})
	for k, v := range e.all {
		exts[k] = v
	}
	return exts
}
