package form

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/pipelined/pipe"

	"github.com/pipelined/phono/input"

	"github.com/pipelined/mp3"
	"github.com/pipelined/signal"
)

// ErrUnsupportedConfig is returned when unsupported configuraton passed.
type ErrUnsupportedConfig string

// Error returns error message.
func (e ErrUnsupportedConfig) Error() string {
	return string(e)
}

// InputMaxSize of file from http request.
func (c Convert) InputMaxSize(r *http.Request) (int64, error) {
	ext := strings.ToLower(path.Base(r.URL.Path))
	switch ext {
	case input.Mp3.DefaultExtension:
		return c.Mp3MaxSize, nil
	case input.Wav.DefaultExtension:
		return c.WavMaxSize, nil
	default:
		return 0, fmt.Errorf("Format %s not supported", ext)
	}
}

// ParsePump returns pump defined as input for conversion.
func (Convert) ParsePump(r *http.Request) (pipe.Pump, io.Closer, error) {
	f, handler, err := r.FormFile(fileKey)
	if err != nil {
		return nil, nil, fmt.Errorf("Invalid file: %v", err)
	}
	switch {
	case input.HasExtension(handler.Filename, input.Wav.Extensions):
		return input.Wav.Pump(f), f, nil
	case input.HasExtension(handler.Filename, input.Mp3.Extensions):
		return input.Mp3.Pump(f), f, nil
	default:
		if err := f.Close(); err != nil {
			return nil, nil, fmt.Errorf("File has unsupported extension: %v \nFailed to close form file: %v", handler.Filename, err)
		}
		return nil, nil, fmt.Errorf("File has unsupported extension: %v", handler.Filename)
	}
}

// ParseSink provided via form.
// This function should return extensions, sinkbuilder
func (Convert) ParseSink(r *http.Request) (fn input.BuildFunc, ext string, err error) {
	ext = r.FormValue("format")
	switch ext {
	case input.Wav.DefaultExtension:
		fn, err = parseWavSink(r)
	case input.Mp3.DefaultExtension:
		fn, err = parseMp3Sink(r)
	default:
		err = ErrUnsupportedConfig(fmt.Sprintf("Unsupported format: %v", ext))
	}
	return
}

func parseWavSink(r *http.Request) (input.BuildFunc, error) {
	// try to get bit depth
	bitDepth, err := parseIntValue(r, "wav-bit-depth", "bit depth")
	if err != nil {
		return nil, err
	}
	return input.Wav.Build(signal.BitDepth(bitDepth))
}

func parseMp3Sink(r *http.Request) (input.BuildFunc, error) {
	// try to get channel mode
	channelMode, err := parseIntValue(r, "mp3-channel-mode", "channel mode")
	if err != nil {
		return nil, err
	}

	// try to get bit rate mode
	bitRateModeString := r.FormValue("mp3-bit-rate-mode")
	if bitRateModeString == "" {
		return nil, fmt.Errorf("Please provide bit rate mode")
	}

	var bitRateMode mp3.BitRateMode
	switch bitRateModeString {
	case mp3opts.VBR:
		// try to get vbr quality
		vbrQuality, err := parseIntValue(r, "mp3-vbr-quality", "vbr quality")
		if err != nil {
			return nil, err
		}
		bitRateMode = mp3.VBR{
			Quality: vbrQuality,
		}
	case mp3opts.CBR:
		// try to get bitrate
		bitRate, err := parseIntValue(r, "mp3-bit-rate", "bit rate")
		if err != nil {
			return nil, err
		}

		bitRateMode = mp3.CBR{
			BitRate: bitRate,
		}
	case mp3opts.ABR:
		// try to get bitrate
		bitRate, err := parseIntValue(r, "mp3-bit-rate", "bit rate")
		if err != nil {
			return nil, err
		}
		bitRateMode = mp3.ABR{
			BitRate: bitRate,
		}
	}

	// try to get mp3 quality
	useQuality, err := parseBoolValue(r, "mp3-use-quality", "mp3 quality")
	if err != nil {
		return nil, err
	}
	var quality int
	if useQuality {
		quality, err = parseIntValue(r, "mp3-quality", "mp3 quality")
		if err != nil {
			return nil, err
		}
	}

	return input.Mp3.Build(bitRateMode, mp3.ChannelMode(channelMode), useQuality, quality)
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

// parseBoolValue parses value of key provided in the html form.
// Returns false if value is not provided. Returns error when cannot be parsed as bool.
func parseBoolValue(r *http.Request, key, name string) (bool, error) {
	str := r.FormValue(key)
	if str == "" {
		return false, nil
	}

	val, err := strconv.ParseBool(str)
	if err != nil {
		return false, ErrUnsupportedConfig(fmt.Sprintf("Failed parsing %s %s: %v", name, str, err))
	}
	return val, nil
}
