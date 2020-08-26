package form_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pipelined/phono/encode/form"
)

func TestEncodeForm(t *testing.T) {
	// wavMaxSize := int64(10)
	// mp3MaxSize := int64(15)
	// encodeForm := form.New(form.Limits{
	// 	WAV: wavMaxSize,
	// 	MP3: mp3MaxSize,
	// })

	// test form data
	// d := encodeForm.Bytes()
	// assert.NotNil(t, d)

	// // test file key
	// k := form.FormFileKey
	// assert.NotEqual(t, "", k)

	// // test form max input size
	// var inputSizeTests = []struct {
	// 	format   fileformat.Format
	// 	url      string
	// 	maxSize  int64
	// 	negative bool
	// }{
	// 	{
	// 		format:  fileformat.WAV,
	// 		url:     "test/.wav",
	// 		maxSize: wavMaxSize,
	// 	},
	// 	{
	// 		format:  fileformat.MP3,
	// 		url:     "test/.mp3",
	// 		maxSize: mp3MaxSize,
	// 	},
	// 	{
	// 		format:   nil,
	// 		url:      "test/wav",
	// 		negative: true,
	// 	},
	// }
	// for _, test := range inputSizeTests {
	// 	size := encodeForm.InputMaxSize(test.format)
	// 	if test.negative {
	// 		assert.Equal(t, int64(0), size)
	// 	} else {
	// 		assert.Equal(t, test.maxSize, size)
	// 	}
	// }

	var parseSinkTests = []struct {
		values   url.Values
		negative bool
	}{
		{
			values:   nil,
			negative: true,
		},
		{
			values: map[string][]string{
				"format":        {".wav"},
				"wav-bit-depth": {"16"},
			},
		},
		{
			values: map[string][]string{
				"format":            {".mp3"},
				"mp3-channel-mode":  {"1"},
				"mp3-bit-rate-mode": {"VBR"},
				"mp3-vbr-quality":   {"1"},
				"mp3-use-quality":   {"true"},
				"mp3-quality":       {"1"},
			},
		},
		{
			values: map[string][]string{
				"format":            {".mp3"},
				"mp3-channel-mode":  {"1"},
				"mp3-bit-rate-mode": {"CBR"},
				"mp3-bit-rate":      {"320"},
				"mp3-use-quality":   {""},
			},
		},
		{
			values: map[string][]string{
				"format":            {".mp3"},
				"mp3-channel-mode":  {"1"},
				"mp3-bit-rate-mode": {"ABR"},
				"mp3-bit-rate":      {"320"},
				"mp3-use-quality":   {""},
			},
		},
		{
			values: map[string][]string{
				"format": {".wav"},
			},
			negative: true,
		},
		{
			values: map[string][]string{
				"format":        {".wav"},
				"wav-bit-depth": {"16bits"},
			},
			negative: true,
		},
		{
			values: map[string][]string{
				"format":           {".mp3"},
				"mp3-channel-mode": {"100"},
			},
			negative: true,
		},
		{
			values: map[string][]string{
				"format": {".mp3"},
			},
			negative: true,
		},
		{
			values: map[string][]string{
				"format":            {".mp3"},
				"mp3-channel-mode":  {"1"},
				"mp3-bit-rate-mode": {"VBR"},
				"mp3-vbr-quality":   {""},
			},
			negative: true,
		},
		{
			values: map[string][]string{
				"format":            {".mp3"},
				"mp3-channel-mode":  {"1"},
				"mp3-bit-rate-mode": {"VBR"},
				"mp3-vbr-quality":   {"1"},
				"mp3-use-quality":   {"trues"},
			},
			negative: true,
		},
		{
			values: map[string][]string{
				"format":            {".mp3"},
				"mp3-channel-mode":  {"1"},
				"mp3-bit-rate-mode": {"VBR"},
				"mp3-vbr-quality":   {"1"},
				"mp3-use-quality":   {"true"},
				"mp3-quality":       {""},
			},

			negative: true,
		},
		{
			values: map[string][]string{
				"format":            {".mp3"},
				"mp3-channel-mode":  {"1"},
				"mp3-bit-rate-mode": {"CBR"},
				"mp3-bit-rate":      {"320s"},
			},
			negative: true,
		},
		{
			values: map[string][]string{
				"format":            {".mp3"},
				"mp3-channel-mode":  {"1"},
				"mp3-bit-rate-mode": {"ABR"},
				"mp3-bit-rate":      {"320s"},
			},
			negative: true,
		},
	}

	for _, test := range parseSinkTests {
		buildFn, ext, err := form.ParseForm(test.values)
		if test.negative {
			assert.NotNil(t, err)
			assert.Equal(t, "", ext)
			assert.Nil(t, buildFn)
		} else {
			assert.Nil(t, err)
			assert.NotEqual(t, "", ext)
			assert.NotNil(t, buildFn)
		}
	}
}
