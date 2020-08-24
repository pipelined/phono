package input

import (
	"fmt"
	"io"
	"strings"

	"pipelined.dev/audio/mp3"
	"pipelined.dev/audio/wav"
	"pipelined.dev/pipe"
	"pipelined.dev/signal"
)

type (
	wavSink struct {
		BitDepths map[signal.BitDepth]struct{}
	}

	mp3Sink struct {
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

	// Sink is used to inject WriteSeeker into Sink.
	Sink func(io.WriteSeeker) pipe.SinkAllocatorFunc
)

var (
	// WAV provides structures required to handle wav files.
	WAV = wavSink{
		BitDepths: map[signal.BitDepth]struct{}{
			signal.BitDepth8:  {},
			signal.BitDepth16: {},
			signal.BitDepth24: {},
			signal.BitDepth32: {},
		},
	}

	// MP3 provides structures required to handle mp3 files.
	MP3 = mp3Sink{
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
)

// hasExtension validates if filename has one of passed extensions.
// Filename is lower-cased before comparison.
func hasExtension(ext string, exts map[string]struct{}) bool {
	_, ok := exts[ext]
	return ok
}

// WAVSink validates all parameters required to build wav sink. If valid, build closure is returned.
// Closure allows to postpone io opertaions and do them only after all sink parameters are validated.
func (f wavSink) Sink(bitDepth int) (Sink, error) {
	bd := signal.BitDepth(bitDepth)
	if _, ok := f.BitDepths[bd]; !ok {
		return nil, fmt.Errorf("Bit depth %v is not supported", bitDepth)
	}

	return func(ws io.WriteSeeker) pipe.SinkAllocatorFunc {
		return wav.Sink(ws, bd)
	}, nil
}

// Sink validates all parameters required to build mp3 sink. If valid, Sink closure is returned.
// Closure allows to postpone io opertaions and do them only after all sink parameters are validated.
func (f mp3Sink) Sink(bitRateMode string, bitRate, channelMode int, useQuality bool, quality int) (Sink, error) {
	cm := mp3.ChannelMode(channelMode)
	if _, ok := f.ChannelModes[cm]; !ok {
		return nil, fmt.Errorf("Channel mode %v is not supported", cm)
	}

	var brm mp3.BitRateMode
	switch strings.ToUpper(bitRateMode) {
	case f.VBR:
		if bitRate < f.MinVBR || bitRate > f.MaxVBR {
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
		return nil, fmt.Errorf("Bit rate mode %v is not supported", bitRateMode)
	}

	if useQuality {
		if quality < f.MinQuality || quality > f.MaxQuality {
			return nil, fmt.Errorf("MP3 quality %v is not supported", quality)
		}
	}

	return func(ws io.WriteSeeker) pipe.SinkAllocatorFunc {
		eq := mp3.DefaultEncodingQuality
		if useQuality {
			eq = mp3.EncodingQuality(quality)
		}
		return mp3.Sink(ws, brm, cm, eq)
	}, nil
}

// BitRate checks if provided bit rate is supported.
func (f mp3Sink) bitRate(v int) error {
	if v > f.MaxBitRate || v < f.MinBitRate {
		return fmt.Errorf("Bit rate %v is not supported. Provide value between %d and %d", v, f.MinBitRate, f.MaxBitRate)
	}
	return nil
}
