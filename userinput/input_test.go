package userinput_test

import (
	"testing"

	"github.com/pipelined/phono/userinput"
	"github.com/stretchr/testify/assert"
)

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
		sinkFn, err := userinput.WAV.Sink(test.bitDepth)
		if test.negative {
			assert.NotNil(t, err)
			assert.Nil(t, sinkFn)
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, sinkFn)
			sink := sinkFn(nil)
			assert.NotNil(t, sink)
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
			bitRateMode: userinput.MP3.VBR,
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
			bitRateMode: userinput.MP3.CBR,
			bitRate:     320,
			channelMode: 2,
		},
		{
			bitRateMode: userinput.MP3.ABR,
			bitRate:     192,
			channelMode: 1,
			useQuality:  true,
		},
		{
			bitRateMode: "fake",
			negative:    true,
		},
		{
			bitRateMode: userinput.MP3.VBR,
			bitRate:     10,
			negative:    true,
		},
		{
			bitRateMode: userinput.MP3.CBR,
			bitRate:     1,
			negative:    true,
		},
		{
			bitRateMode: userinput.MP3.ABR,
			bitRate:     1,
			negative:    true,
		},
		{
			bitRateMode: userinput.MP3.VBR,
			bitRate:     1,
			channelMode: 100,
			negative:    true,
		},
		{
			bitRateMode: userinput.MP3.VBR,
			bitRate:     1,
			channelMode: 100,
			negative:    true,
		},
		{
			bitRateMode: userinput.MP3.VBR,
			bitRate:     1,
			channelMode: 1,
			useQuality:  true,
			quality:     10,
			negative:    true,
		},
	}
	for _, test := range tests {
		sinkFn, err := userinput.MP3.Sink(
			test.bitRateMode,
			test.bitRate,
			test.channelMode,
			test.useQuality,
			test.quality,
		)
		if test.negative {
			assert.NotNil(t, err)
			assert.Nil(t, sinkFn)
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, sinkFn)
			sink := sinkFn(nil)
			assert.NotNil(t, sink)
		}
	}
}
