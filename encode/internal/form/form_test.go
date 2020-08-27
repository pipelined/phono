package form_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"golang.org/x/net/html"

	"github.com/pipelined/phono/encode/internal/form"
	"pipelined.dev/audio/fileformat"
)

func TestFormParsing(t *testing.T) {
	newRequest := func(uri, filePath string, params map[string]string) *http.Request {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		if filePath != "" {
			file, err := os.Open(filePath)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			part, err := writer.CreateFormFile(form.FormFileKey, filepath.Base(filePath))
			if err != nil {
				panic(err)
			}
			_, err = io.Copy(part, file)
		}

		for key, val := range params {
			_ = writer.WriteField(key, val)
		}
		err := writer.Close()
		if err != nil {
			panic(err)
		}

		req, err := http.NewRequest("POST", uri, body)
		if err != nil {
			panic(err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
		return req
	}

	newWavRequest := func(params map[string]string) *http.Request {
		return newRequest("test/.wav", "../../../_testdata/sample.wav", params)
	}

	testOk := func(f form.Form, r *http.Request) func(*testing.T) {
		return func(t *testing.T) {
			_, err := f.Parse(r)
			assertEqual(t, "error", err, nil)
		}
	}
	testFail := func(f form.Form, r *http.Request) func(*testing.T) {
		return func(t *testing.T) {
			_, err := f.Parse(r)
			assertNotNil(t, "error", err)
		}
	}

	noLimits := form.Limits{}
	t.Run("ok wav",
		testOk(form.New(noLimits),
			newWavRequest(
				map[string]string{
					"format":        ".wav",
					"wav-bit-depth": "16",
				},
			),
		),
	)
	t.Run("ok mp3 vbr",
		testOk(form.New(noLimits),
			newWavRequest(
				map[string]string{
					"format":            ".mp3",
					"mp3-channel-mode":  "1",
					"mp3-bit-rate-mode": "VBR",
					"mp3-vbr-quality":   "1",
					"mp3-use-quality":   "true",
					"mp3-quality":       "1",
				},
			),
		),
	)
	t.Run("ok mp3 cbr",
		testOk(form.New(noLimits),
			newWavRequest(map[string]string{
				"format":            ".mp3",
				"mp3-channel-mode":  "1",
				"mp3-bit-rate-mode": "CBR",
				"mp3-bit-rate":      "320",
			},
			),
		),
	)
	t.Run("ok mp3 abr",
		testOk(form.New(noLimits),
			newWavRequest(map[string]string{
				"format":            ".mp3",
				"mp3-channel-mode":  "1",
				"mp3-bit-rate-mode": "ABR",
				"mp3-bit-rate":      "320",
			}),
		),
	)
	t.Run("fail size exceeded",
		testFail(form.New(form.Limits{fileformat.WAV: 10}),
			newWavRequest(nil),
		),
	)
	t.Run("fail input format",
		testFail(form.New(form.Limits{fileformat.WAV: 10}),
			newRequest("non-existing-format", "", nil),
		),
	)
	t.Run("fail output format",
		testFail(form.New(form.Limits{fileformat.WAV: 10}),
			newWavRequest(map[string]string{
				"format": "non-existing-format",
			}),
		),
	)
	t.Run("fail no file",
		testFail(form.New(form.Limits{fileformat.WAV: 10}),
			newRequest(".wav", "", nil),
		),
	)
	t.Run("fail wav missing bit depth",
		testFail(form.New(form.Limits{}),
			newWavRequest(map[string]string{
				"format":        ".wav",
				"wav-bit-depth": "",
			})),
	)
	t.Run("fail mp3 invalid channel mode",
		testFail(form.New(form.Limits{}),
			newWavRequest(map[string]string{
				"format":           ".mp3",
				"mp3-channel-mode": "invalid-channel-mode",
			}),
		),
	)
	t.Run("fail mp3 invalid bit rate mode",
		testFail(form.New(form.Limits{}),
			newWavRequest(map[string]string{
				"format":            ".mp3",
				"mp3-channel-mode":  "1",
				"mp3-bit-rate-mode": "invalid-bit-rate-mode",
			}),
		),
	)
	t.Run("fail mp3 invalid vbr quality",
		testFail(form.New(form.Limits{}),
			newWavRequest(map[string]string{
				"format":            ".mp3",
				"mp3-channel-mode":  "1",
				"mp3-bit-rate-mode": "VBR",
				"mp3-vbr-quality":   "",
			}),
		),
	)
	t.Run("fail mp3 invalid bit rate",
		testFail(form.New(form.Limits{}),
			newWavRequest(map[string]string{
				"format":            ".mp3",
				"mp3-channel-mode":  "1",
				"mp3-bit-rate-mode": "CBR",
				"mp3-bit-rate":      "",
			}),
		),
	)
	t.Run("fail mp3 invalid quality flag",
		testFail(form.New(form.Limits{}),
			newWavRequest(map[string]string{
				"format":            ".mp3",
				"mp3-channel-mode":  "1",
				"mp3-bit-rate-mode": "CBR",
				"mp3-bit-rate":      "320",
				"mp3-use-quality":   "no",
			}),
		),
	)
	t.Run("fail mp3 invalid quality value",
		testFail(form.New(form.Limits{}),
			newWavRequest(map[string]string{
				"format":            ".mp3",
				"mp3-channel-mode":  "1",
				"mp3-bit-rate-mode": "CBR",
				"mp3-bit-rate":      "320",
				"mp3-use-quality":   "true",
				"mp3-quality":       "non-int",
			}),
		),
	)
}

func TestForm(t *testing.T) {
	f := form.New(form.Limits{})
	_, err := html.Parse(bytes.NewReader(f.Bytes()))
	assertEqual(t, "html error", err, nil)
}

func assertEqual(t *testing.T, name string, result, expected interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("%v\nresult: \t%T\t%+v \nexpected: \t%T\t%+v", name, result, result, expected, expected)
	}
}

func assertNotNil(t *testing.T, name string, result interface{}) {
	t.Helper()
	if reflect.DeepEqual(nil, result) {
		t.Fatalf("%v\nresult: \t%T\t%+v \nexpected: \t%T\t%+v", name, result, result, nil, nil)
	}
}

func assertPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()
	fn()
}
