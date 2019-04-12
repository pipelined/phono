package template

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/pipelined/mp3"
	"github.com/pipelined/phono/convert"
	"github.com/pipelined/pipe"
	"github.com/pipelined/signal"
	"github.com/pipelined/wav"
)

// ErrUnsupportedConfig is returned when unsupported configuraton passed.
type ErrUnsupportedConfig string

// Error returns error message.
func (e ErrUnsupportedConfig) Error() string {
	return string(e)
}

// ConvertForm provides user interaction via http form.
type ConvertForm struct{}

// Data returns serialized form data, ready to be served.
func (ConvertForm) Data() []byte {
	return convertFormData
}

// Format parses input format from http request.
func (ConvertForm) Format(r *http.Request) string {
	return path.Base(r.URL.Path)
}

// ParsePump returns pump defined as input for conversion.
func (ConvertForm) ParsePump(r *http.Request) (pipe.Pump, io.Closer, error) {
	f, handler, err := r.FormFile("input-file")
	if err != nil {
		return nil, nil, fmt.Errorf("Invalid file: %v", err)
	}
	switch {
	case hasExtension(handler.Filename, wav.Extensions):
		return &wav.Pump{ReadSeeker: f}, f, nil
	case hasExtension(handler.Filename, mp3.Extensions):
		return &mp3.Pump{Reader: f}, f, nil
	default:
		extErr := fmt.Errorf("File has unsupported extension: %v", handler.Filename)
		if err = f.Close(); err != nil {
			return nil, nil, fmt.Errorf("%s \nFailed close form file: %v", extErr, err)
		}
		return nil, nil, extErr
	}
}

// ParseOutput data provided via form.
// This function should return extensions, sinkbuilder
func (ConvertForm) ParseOutput(r *http.Request) (sb convert.SinkBuilder, ext string, err error) {
	ext = r.FormValue("format")
	switch ext {
	case wav.DefaultExtension:
		sb, err = parseWavConfig(r)
		return
	case mp3.DefaultExtension:
		sb, err = parseMp3Config(r)
		return
	default:
		return nil, "", ErrUnsupportedConfig(fmt.Sprintf("Unsupported format: %v", ext))
	}
}

func parseWavConfig(r *http.Request) (*wav.SinkBuilder, error) {
	// try to get bit depth
	bitDepth, err := parseIntValue(r, "wav-bit-depth", "bit depth")
	if err != nil {
		return nil, err
	}

	return &wav.SinkBuilder{BitDepth: signal.BitDepth(bitDepth)}, nil
}

func parseMp3Config(r *http.Request) (*mp3.SinkBuilder, error) {
	// try to get bit rate mode
	bitRateMode, err := parseIntValue(r, "mp3-bit-rate-mode", "bit rate mode")
	if err != nil {
		return nil, err
	}

	// try to get channel mode
	channelMode, err := parseIntValue(r, "mp3-channel-mode", "channel mode")
	if err != nil {
		return nil, err
	}

	if mp3.BitRateMode(bitRateMode) == mp3.VBR {
		// try to get vbr quality
		vbrQuality, err := parseIntValue(r, "mp3-vbr-quality", "vbr quality")
		if err != nil {
			return nil, err
		}
		return &mp3.SinkBuilder{
			BitRateMode: mp3.VBR,
			ChannelMode: mp3.ChannelMode(channelMode),
			VBRQuality:  vbrQuality,
		}, nil
	}

	// try to get bitrate
	bitRate, err := parseIntValue(r, "mp3-bit-rate", "bit rate")
	if err != nil {
		return nil, err
	}
	return &mp3.SinkBuilder{
		BitRateMode: mp3.BitRateMode(bitRateMode),
		ChannelMode: mp3.ChannelMode(channelMode),
		BitRate:     bitRate,
	}, nil
}

// parseIntValue parses value of key provided in the html form.
// Returns error if value is not provided or cannot be parsed as int.
func parseIntValue(r *http.Request, key, name string) (int, error) {
	str := r.FormValue(key)
	if str == "" {
		return 0, ErrUnsupportedConfig(fmt.Sprintf("Please provide %s", name))
	}

	val, err := strconv.Atoi(str)
	if err != nil {
		return 0, ErrUnsupportedConfig(fmt.Sprintf("Failed parsing %s %s: %v", name, str, err))
	}
	return val, nil
}

func hasExtension(fileName string, fn extensionsFunc) bool {
	for _, ext := range fn() {
		if strings.HasSuffix(strings.ToLower(fileName), ext) {
			return true
		}
	}
	return false
}
