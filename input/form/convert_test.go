package form_test

import (
	"net/url"
	"testing"

	"github.com/pipelined/phono/input/form"
	"github.com/stretchr/testify/assert"
)

func TestConvertForm(t *testing.T) {
	wavMaxSize := int64(10)
	mp3MaxSize := int64(15)
	convertForm := form.Convert{
		WavMaxSize: wavMaxSize,
		Mp3MaxSize: mp3MaxSize,
	}

	// test form data
	d := convertForm.Data()
	assert.NotNil(t, d)

	// test file key
	k := convertForm.FileKey()
	assert.NotEqual(t, "", k)

	// test form max input size
	var inputSizeTests = []struct {
		url      string
		maxSize  int64
		negative bool
	}{
		{
			url:     "test/.wav",
			maxSize: wavMaxSize,
		},
		{
			url:     "test/.mp3",
			maxSize: mp3MaxSize,
		},
		{
			url:      "test/wav",
			negative: true,
		},
		{
			url:      "test/mp3",
			negative: true,
		},
	}
	for _, test := range inputSizeTests {
		size, err := convertForm.InputMaxSize(test.url)
		if test.negative {
			assert.NotNil(t, err)
			assert.Equal(t, int64(0), size)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, test.maxSize, size)
		}
	}

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
				"format":        []string{".wav"},
				"wav-bit-depth": []string{"16"},
			},
		},
		{
			values: map[string][]string{
				"format":            []string{".mp3"},
				"mp3-channel-mode":  []string{"1"},
				"mp3-bit-rate-mode": []string{"VBR"},
				"mp3-vbr-quality":   []string{"1"},
				"mp3-use-quality":   []string{"true"},
				"mp3-quality":       []string{"1"},
			},
		},
		{
			values: map[string][]string{
				"format":            []string{".mp3"},
				"mp3-channel-mode":  []string{"1"},
				"mp3-bit-rate-mode": []string{"CBR"},
				"mp3-bit-rate":      []string{"320"},
				"mp3-use-quality":   []string{""},
			},
		},
		{
			values: map[string][]string{
				"format":            []string{".mp3"},
				"mp3-channel-mode":  []string{"1"},
				"mp3-bit-rate-mode": []string{"ABR"},
				"mp3-bit-rate":      []string{"320"},
				"mp3-use-quality":   []string{""},
			},
		},
		{
			values: map[string][]string{
				"format": []string{".wav"},
			},
			negative: true,
		},
		{
			values: map[string][]string{
				"format":        []string{".wav"},
				"wav-bit-depth": []string{"16bits"},
			},
			negative: true,
		},
		{
			values: map[string][]string{
				"format":           []string{".mp3"},
				"mp3-channel-mode": []string{"100"},
			},
			negative: true,
		},
		{
			values: map[string][]string{
				"format": []string{".mp3"},
			},
			negative: true,
		},
		{
			values: map[string][]string{
				"format":            []string{".mp3"},
				"mp3-channel-mode":  []string{"1"},
				"mp3-bit-rate-mode": []string{"VBR"},
				"mp3-vbr-quality":   []string{""},
			},
			negative: true,
		},
		{
			values: map[string][]string{
				"format":            []string{".mp3"},
				"mp3-channel-mode":  []string{"1"},
				"mp3-bit-rate-mode": []string{"VBR"},
				"mp3-vbr-quality":   []string{"1"},
				"mp3-use-quality":   []string{"trues"},
			},
			negative: true,
		},
		{
			values: map[string][]string{
				"format":            []string{".mp3"},
				"mp3-channel-mode":  []string{"1"},
				"mp3-bit-rate-mode": []string{"VBR"},
				"mp3-vbr-quality":   []string{"1"},
				"mp3-use-quality":   []string{"true"},
				"mp3-quality":       []string{""},
			},

			negative: true,
		},
		{
			values: map[string][]string{
				"format":            []string{".mp3"},
				"mp3-channel-mode":  []string{"1"},
				"mp3-bit-rate-mode": []string{"CBR"},
				"mp3-bit-rate":      []string{"320s"},
			},
			negative: true,
		},
		{
			values: map[string][]string{
				"format":            []string{".mp3"},
				"mp3-channel-mode":  []string{"1"},
				"mp3-bit-rate-mode": []string{"ABR"},
				"mp3-bit-rate":      []string{"320s"},
			},
			negative: true,
		},
	}

	for _, test := range parseSinkTests {
		buildFn, ext, err := convertForm.ParseSink(test.values)
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
