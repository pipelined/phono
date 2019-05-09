// Package input provides types to parse user input of pipe components.
package input

import (
	"io"
	"net/http"

	"github.com/pipelined/mp3"
	"github.com/pipelined/pipe"
	"github.com/pipelined/wav"
)

var (
	DefaultExtension = struct {
		Wav string
		Mp3 string
	}{
		Wav: ".wav",
		Mp3: ".mp3",
	}

	Extensions = struct {
		Wav []string
		Mp3 []string
	}{
		Wav: []string{
			DefaultExtension.Wav, ".wave",
		},
		Mp3: []string{
			DefaultExtension.Mp3,
		},
	}
)

type (
	// Pump contains all pipe.Pumps that user can provide as input.
	Pump struct {
		Wav *wav.Pump
		Mp3 *mp3.Pump
	}

	// Sink contains all pipe.Sinks that user can provide as input.
	Sink struct {
		Mp3 *mp3.Sink
		Wav *wav.Sink
	}

	// ConvertForm provides html form to the user. The form contains all information needed for conversion.
	ConvertForm interface {
		Data() []byte
		InputMaxSize(r *http.Request) (int64, error)
		ParsePump(r *http.Request) (Pump, io.Closer, error)
		ParseSink(r *http.Request) (Sink, error)
	}
)

func (s Sink) mp3() bool {
	return s.Mp3 != nil
}

func (s Sink) wav() bool {
	return s.Wav != nil
}

func (p Pump) mp3() bool {
	return p.Mp3 != nil
}

func (p Pump) wav() bool {
	return p.Wav != nil
}

// SetOutput to the sink provided as input.
func (s Sink) SetOutput(ws io.WriteSeeker) {
	switch {
	case s.mp3():
		s.Mp3.Writer = ws
	case s.wav():
		s.Wav.WriteSeeker = ws
	}
}

// Extension of the file for the sink.
func (s Sink) Extension() string {
	switch {
	case s.mp3():
		return DefaultExtension.Mp3
	case s.wav():
		return DefaultExtension.Wav
	default:
		return ""
	}
}

// Sink provided as input.
func (s Sink) Sink() pipe.Sink {
	switch {
	case s.mp3():
		return s.Mp3
	case s.wav():
		return s.Wav
	default:
		return nil
	}
}

// Pump provided as input.
func (p Pump) Pump() pipe.Pump {
	switch {
	case p.mp3():
		return p.Mp3
	case p.wav():
		return p.Wav
	default:
		return nil
	}
}
