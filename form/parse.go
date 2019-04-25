package form

import (
	"fmt"
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

// FileKey returns the
func (Convert) FileKey() string {
	return fileKey
}

// ParseExtension of input file from http request.
func (Convert) ParseExtension(r *http.Request) string {
	return fmt.Sprintf(".%s", path.Base(r.URL.Path))
}

// ParsePump returns pump defined as input for conversion.
func (Convert) ParsePump(fileName string) (pipe.Pump, error) {
	switch {
	case hasExtension(fileName, wav.Extensions):
		return &wav.Pump{}, nil
	case hasExtension(fileName, mp3.Extensions):
		return &mp3.Pump{}, nil
	default:
		return nil, fmt.Errorf("File has unsupported extension: %v", fileName)
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
	if err := wav.Supported.BitDepth(signal.BitDepth(bitDepth)); err != nil {
		return nil, fmt.Errorf("Bit depth %v is not supported", bitDepth)
	}

	return &wav.Sink{
		BitDepth: signal.BitDepth(bitDepth),
	}, nil
}

func parseMp3Sink(r *http.Request) (*mp3.Sink, error) {
	// try to get bit rate mode
	bitRateModeString := r.FormValue("mp3-bit-rate-mode")
	if bitRateModeString == "" {
		return nil, fmt.Errorf("Please provide bit rate mode")
	}

	// try to get channel mode
	channelMode, err := parseIntValue(r, "mp3-channel-mode", "channel mode")
	if err != nil {
		return nil, err
	}
	if err := mp3.Supported.ChannelMode(mp3.ChannelMode(channelMode)); err != nil {
		return nil, fmt.Errorf("Channel mode %v is not supported", channelMode)
	}

	var bitRateMode mp3.BitRateMode
	switch bitRateModeString {
	case mp3.VBR{}.String():
		// try to get vbr quality
		vbrQuality, err := parseIntValue(r, "mp3-vbr-quality", "vbr quality")
		if err != nil {
			return nil, err
		}
		if vbrQuality < 0 || vbrQuality > 9 {
			return nil, fmt.Errorf("VBR quality %v is not supported", vbrQuality)
		}

		bitRateMode = mp3.VBR{
			Quality: vbrQuality,
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

		if bitRateModeString == (mp3.CBR{}.String()) {
			bitRateMode = mp3.CBR{
				BitRate: bitRate,
			}
		} else {
			bitRateMode = mp3.ABR{
				BitRate: bitRate,
			}
		}
	}

	s := mp3.Sink{
		BitRateMode: bitRateMode,
		ChannelMode: mp3.ChannelMode(channelMode),
	}

	// try to get mp3 quality
	useQuality, err := parseBoolValue(r, "mp3-use-quality", "mp3 quality")
	if err != nil {
		return nil, err
	}
	if useQuality {
		mp3Quality, err := parseIntValue(r, "mp3-quality", "mp3 quality")
		if err != nil {
			return nil, err
		}
		if mp3Quality < 0 || mp3Quality > 9 {
			return nil, fmt.Errorf("MP3 quality %v is not supported", mp3Quality)
		}
		s.SetQuality(mp3Quality)
	}

	return &s, nil
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

func hasExtension(fileName string, fn extensionsFunc) bool {
	for _, ext := range fn() {
		if strings.HasSuffix(strings.ToLower(fileName), ext) {
			return true
		}
	}
	return false
}
