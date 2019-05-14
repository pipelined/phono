package form

import (
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/pipelined/phono/input"
)

// ErrUnsupportedConfig is returned when unsupported configuraton passed.
type ErrUnsupportedConfig string

// Error returns error message.
func (e ErrUnsupportedConfig) Error() string {
	return string(e)
}

// InputMaxSize of file from http request.
func (c Convert) InputMaxSize(url string) (int64, error) {
	ext := strings.ToLower(path.Base(url))
	switch ext {
	case input.Mp3.DefaultExtension:
		return c.Mp3MaxSize, nil
	case input.Wav.DefaultExtension:
		return c.WavMaxSize, nil
	default:
		return 0, fmt.Errorf("Format %s not supported", ext)
	}
}

// FileKey returns a name of form file value.
func (Convert) FileKey() string {
	return fileKey
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
	return input.Wav.Build(bitDepth)
}

func parseMp3Sink(r *http.Request) (input.BuildFunc, error) {
	// try to get channel mode
	channelMode, err := parseIntValue(r, "mp3-channel-mode", "channel mode")
	if err != nil {
		return nil, err
	}

	// try to get bit rate mode
	bitRateMode := r.FormValue("mp3-bit-rate-mode")
	if bitRateMode == "" {
		return nil, fmt.Errorf("Please provide bit rate mode")
	}

	var bitRate int
	switch bitRateMode {
	case input.Mp3.VBR:
		// try to get vbr quality
		bitRate, err = parseIntValue(r, "mp3-vbr-quality", "vbr quality")
		if err != nil {
			return nil, err
		}
	case input.Mp3.CBR:
		// try to get bitrate
		bitRate, err = parseIntValue(r, "mp3-bit-rate", "bit rate")
		if err != nil {
			return nil, err
		}
	case input.Mp3.ABR:
		// try to get bitrate
		bitRate, err = parseIntValue(r, "mp3-bit-rate", "bit rate")
		if err != nil {
			return nil, err
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

	return input.Mp3.Build(bitRateMode, bitRate, channelMode, useQuality, quality)
}

// parseIntValue parses value of key provided in the html form.
// Returns error if value is not provided or cannot be parsed as int.
func parseIntValue(r *http.Request, key, name string) (int, error) {
	str := r.FormValue(key)
	if str == "" {
		return 0, ErrUnsupportedConfig(fmt.Sprintf("%s not provided", name))
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
