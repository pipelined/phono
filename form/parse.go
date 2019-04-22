package form

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/pipelined/mp3"
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

// Convert provides user interaction via http form.
type Convert struct{}

// Data returns serialized form data, ready to be served.
func (Convert) Data() []byte {
	return convertFormData
}

// ParseExtension of input file from http request.
func (Convert) ParseExtension(r *http.Request) string {
	return fmt.Sprintf(".%s", path.Base(r.URL.Path))
}

// ParsePump returns pump defined as input for conversion.
func (Convert) ParsePump(r *http.Request) (pipe.Pump, io.Closer, error) {
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
func (Convert) ParseOutput(r *http.Request) (s pipe.Sink, ext string, err error) {
	ext = r.FormValue("format")
	switch ext {
	case wav.DefaultExtension:
		s, err = parseWavSink(r)
		return
	case mp3.DefaultExtension:
		s, err = parseMp3Sink(r)
		return
	default:
		return nil, "", ErrUnsupportedConfig(fmt.Sprintf("Unsupported format: %v", ext))
	}
}

func parseWavSink(r *http.Request) (*wav.Sink, error) {
	// try to get bit depth
	bitDepth, err := parseIntValue(r, "wav-bit-depth", "bit depth")
	if err != nil {
		return nil, err
	}
	if _, ok := wav.Supported.BitDepths[signal.BitDepth(bitDepth)]; !ok {
		return nil, fmt.Errorf("Bit depth %v is not supported", bitDepth)
	}

	return &wav.Sink{BitDepth: signal.BitDepth(bitDepth)}, nil
}

func parseMp3Sink(r *http.Request) (mp3.Sink, error) {
	// try to get bit rate mode
	bitRateMode, err := parseIntValue(r, "mp3-bit-rate-mode", "bit rate mode")
	if err != nil {
		return nil, err
	}
	if _, ok := mp3.Supported.BitRateModes[mp3.BitRateMode(bitRateMode)]; !ok {
		return nil, fmt.Errorf("Bit rate mode %v is not supported", bitRateMode)
	}

	// try to get channel mode
	channelMode, err := parseIntValue(r, "mp3-channel-mode", "channel mode")
	if err != nil {
		return nil, err
	}
	if _, ok := mp3.Supported.ChannelModes[mp3.ChannelMode(channelMode)]; !ok {
		return nil, fmt.Errorf("Channel mode %v is not supported", channelMode)
	}

	var s mp3.Sink
	switch mp3.BitRateMode(bitRateMode) {
	case mp3.VBR:
		// try to get vbr quality
		vbrQuality, err := parseIntValue(r, "mp3-vbr-quality", "vbr quality")
		if err != nil {
			return nil, err
		}
		if vbrQuality < 0 || vbrQuality > 9 {
			return nil, fmt.Errorf("VBR quality %v is not supported", vbrQuality)
		}

		s = &mp3.VBRSink{
			ChannelMode: mp3.ChannelMode(channelMode),
			VBRQuality:  vbrQuality,
		}
	default:
		// try to get bitrate
		bitRate, err := parseIntValue(r, "mp3-bit-rate", "bit rate")
		if err != nil {
			return nil, err
		}
		if bitRate > mp3.MaxBitRate || bitRate < mp3.MinBitRate {
			return nil, fmt.Errorf("Bit rate %v is not supported. Provide value between %d and %d", bitRate, mp3.MinBitRate, mp3.MaxBitRate)
		}

		if mp3.BitRateMode(bitRateMode) == mp3.CBR {
			s = &mp3.CBRSink{
				ChannelMode: mp3.ChannelMode(channelMode),
				BitRate:     bitRate,
			}
		} else {
			s = &mp3.ABRSink{
				ChannelMode: mp3.ChannelMode(channelMode),
				BitRate:     bitRate,
			}
		}
	}

	return s, nil
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
