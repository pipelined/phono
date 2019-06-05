package file_test

import (
	"testing"

	"github.com/pipelined/phono/file"
	"github.com/stretchr/testify/assert"
)

func TestFilePump(t *testing.T) {
	var tests = []struct {
		fileName string
		negative bool
	}{
		{
			fileName: "test.wav",
		},
		{
			fileName: "test.mp3",
		},
		{
			fileName: "",
			negative: true,
		},
	}

	for _, test := range tests {
		buildPump, err := file.Pump(test.fileName)
		if test.negative {
			assert.NotNil(t, err)
		} else {
			assert.NotNil(t, buildPump)
		}
		pump := buildPump(nil)
		assert.NotNil(t, pump)
	}
}

func TestBuildWav(t *testing.T) {
	var tests = []struct {
		bitDepth int
		negative bool
	}{
		{
			bitDepth: 16,
		},
		{
			bitDepth: 20,
			negative: true,
		},
	}
	for _, test := range tests {
		buildFn, err := file.Wav.BuildSink(test.bitDepth)
		if test.negative {
			assert.NotNil(t, err)
			assert.Nil(t, buildFn)
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, buildFn)
			pump := buildFn(nil)
			assert.NotNil(t, pump)
		}
	}
}

func TestBuildMp3(t *testing.T) {
	var tests = []struct {
		bitRateMode string
		channelMode int
		bitRate     int
		useQuality  bool
		quality     int
		negative    bool
	}{
		{
			bitRateMode: file.Mp3.VBR,
			bitRate:     1,
			channelMode: 1,
			useQuality:  true,
		},
		{
			bitRateMode: "vbr",
			bitRate:     1,
			channelMode: 1,
			useQuality:  true,
		},
		{
			bitRateMode: file.Mp3.CBR,
			bitRate:     320,
			channelMode: 2,
		},
		{
			bitRateMode: file.Mp3.ABR,
			bitRate:     192,
			channelMode: 1,
			useQuality:  true,
		},
		{
			bitRateMode: "fake",
			negative:    true,
		},
		{
			bitRateMode: file.Mp3.VBR,
			bitRate:     10,
			negative:    true,
		},
		{
			bitRateMode: file.Mp3.CBR,
			bitRate:     1,
			negative:    true,
		},
		{
			bitRateMode: file.Mp3.ABR,
			bitRate:     1,
			negative:    true,
		},
		{
			bitRateMode: file.Mp3.VBR,
			bitRate:     1,
			channelMode: 100,
			negative:    true,
		},
		{
			bitRateMode: file.Mp3.VBR,
			bitRate:     1,
			channelMode: 100,
			negative:    true,
		},
		{
			bitRateMode: file.Mp3.VBR,
			bitRate:     1,
			channelMode: 1,
			useQuality:  true,
			quality:     10,
			negative:    true,
		},
	}
	for _, test := range tests {
		buildFn, err := file.Mp3.BuildSink(
			test.bitRateMode,
			test.bitRate,
			test.channelMode,
			test.useQuality,
			test.quality,
		)
		if test.negative {
			assert.NotNil(t, err)
			assert.Nil(t, buildFn)
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, buildFn)
			pump := buildFn(nil)
			assert.NotNil(t, pump)
		}
	}
}
