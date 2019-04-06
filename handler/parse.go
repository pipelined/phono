package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/pipelined/mp3"
	"github.com/pipelined/phono/convert"
	"github.com/pipelined/signal"
)

func parseConfig(r *http.Request) (convert.OutputConfig, error) {
	f := convert.Format(r.FormValue("format"))
	switch f {
	case convert.WavFormat:
		return parseWavConfig(r)
	case convert.Mp3Format:
		return parseMp3Config(r)
	default:
		return nil, fmt.Errorf("Unsupported format: %v", f)
	}
}

func parseWavConfig(r *http.Request) (convert.WavConfig, error) {
	// try to get bit depth
	bitDepth, err := parseIntValue(r, "wav-bit-depth", "bit depth")
	if err != nil {
		return convert.WavConfig{}, err
	}

	return convert.WavConfig{BitDepth: signal.BitDepth(bitDepth)}, nil
}

func parseMp3Config(r *http.Request) (convert.Mp3Config, error) {
	// try to get bit rate mode
	bitRateMode, err := parseIntValue(r, "mp3-bit-rate-mode", "bit rate mode")
	if err != nil {
		return convert.Mp3Config{}, err
	}

	// try to get channel mode
	channelMode, err := parseIntValue(r, "mp3-channel-mode", "channel mode")
	if err != nil {
		return convert.Mp3Config{}, err
	}

	if mp3.BitRateMode(bitRateMode) == mp3.VBR {
		// try to get vbr quality
		vbrQuality, err := parseIntValue(r, "mp3-vbr-quality", "vbr quality")
		if err != nil {
			return convert.Mp3Config{}, err
		}
		return convert.Mp3Config{
			BitRateMode: mp3.VBR,
			ChannelMode: mp3.ChannelMode(channelMode),
			VBRQuality:  mp3.VBRQuality(vbrQuality),
		}, nil
	}

	// try to get bitrate
	bitRate, err := parseIntValue(r, "mp3-bit-rate", "bit rate")
	if err != nil {
		return convert.Mp3Config{}, err
	}
	return convert.Mp3Config{
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
		return 0, fmt.Errorf("Please provide %s", name)
	}

	val, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("Failed parsing %s %s: %v", name, str, err)
	}
	return val, nil
}
