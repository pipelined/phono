package encode_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pipelined/phono/encode"
	"github.com/pipelined/phono/encode/form"
)

func parseURL(raw string) (result *url.URL) {
	result, _ = url.Parse(raw)
	return
}

// Creates a new file upload http request with optional extra params. Any error causes panic.
func fileUploadRequest(uri string, params map[string]string, filePath string) *http.Request {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(form.FormFileKey, filepath.Base(filePath))
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
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

func wavUploadRequest(params map[string]string) *http.Request {
	return fileUploadRequest("test/.wav", params, "../_testdata/sample.wav")
}

func mp3UploadRequest(params map[string]string) *http.Request {
	return fileUploadRequest("test/.mp3", params, "../_testdata/sample.mp3")
}

func notMediaUploadRequest(uri string, params map[string]string) *http.Request {
	return fileUploadRequest(uri, params, "../_testdata/not-media")
}

func TestHandler(t *testing.T) {
	bufferSize := 512
	testHandler := func(form form.Form, r *http.Request, expectedStatus int) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()
			h := encode.Handler(form, bufferSize, "")
			assert.NotNil(t, h)

			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, r)
			assert.Equal(t, expectedStatus, rr.Code)
		}
	}
	t.Run("not allowed method",
		testHandler(form.Form{},
			&http.Request{
				Method: http.MethodPut,
				URL:    parseURL("test/.wav"),
			},
			http.StatusMethodNotAllowed),
	)
	t.Run("not allowed input format",
		testHandler(form.Form{},
			&http.Request{
				Method: http.MethodPost,
				URL:    parseURL("test/.test"),
			},
			http.StatusBadRequest),
	)
	t.Run("not allowed output format",
		testHandler(form.Form{},
			fileUploadRequest(
				"test/.mp3",
				map[string]string{
					"format": "non-existing-format",
				},
				"../_testdata/sample.mp3",
			),
			http.StatusBadRequest),
	)
	t.Run("wav empty body",
		testHandler(form.Form{},
			&http.Request{
				Method: http.MethodPost,
				URL:    parseURL("test/.wav"),
			},
			http.StatusBadRequest),
	)
	t.Run("wav missing bit depth",
		testHandler(form.Form{},
			wavUploadRequest(nil),
			http.StatusBadRequest),
	)
	t.Run("wav max size exceeded",
		testHandler(form.Form{WavMaxSize: 10},
			wavUploadRequest(map[string]string{
				"format":        ".wav",
				"wav-bit-depth": "16",
			}),
			http.StatusBadRequest),
	)
	t.Run("wav max size exceeded",
		testHandler(form.Form{WavMaxSize: 10},
			wavUploadRequest(map[string]string{
				"format":        ".wav",
				"wav-bit-depth": "16",
			}),
			http.StatusBadRequest),
	)
	t.Run("wav not media type",
		testHandler(form.Form{},
			notMediaUploadRequest("test/.wav", map[string]string{
				"format":        ".wav",
				"wav-bit-depth": "16",
			}),
			http.StatusBadRequest),
	)
	t.Run("wav ok",
		testHandler(form.Form{},
			wavUploadRequest(map[string]string{
				"format":        ".wav",
				"wav-bit-depth": "16",
			}),
			http.StatusOK),
	)
	t.Run("mp3 vbr ok",
		testHandler(form.Form{},
			mp3UploadRequest(map[string]string{
				"format":            ".mp3",
				"mp3-channel-mode":  "1",
				"mp3-bit-rate-mode": "VBR",
				"mp3-vbr-quality":   "1",
				"mp3-use-quality":   "true",
				"mp3-quality":       "1",
			}),
			http.StatusOK),
	)
	t.Run("mp3 cbr ok",
		testHandler(form.Form{},
			mp3UploadRequest(map[string]string{
				"format":            ".mp3",
				"mp3-channel-mode":  "1",
				"mp3-bit-rate-mode": "CBR",
				"mp3-bit-rate":      "320",
			}),
			http.StatusOK),
	)
	t.Run("mp3 abr ok",
		testHandler(form.Form{},
			mp3UploadRequest(map[string]string{
				"format":            ".mp3",
				"mp3-channel-mode":  "1",
				"mp3-bit-rate-mode": "ABR",
				"mp3-bit-rate":      "320",
			}),
			http.StatusOK),
	)
	t.Run("wav as mp3",
		testHandler(form.Form{},
			fileUploadRequest(
				"test/.mp3",
				map[string]string{
					"format": ".wav",
				},
				"../_testdata/sample.wav",
			),
			http.StatusBadRequest),
	)
	t.Run("wav invalid bit depth",
		testHandler(form.Form{},
			wavUploadRequest(
				map[string]string{
					"format":        ".wav",
					"wav-bit-depth": "non-int-value",
				},
			),
			http.StatusBadRequest),
	)
	t.Run("mp3 invalid channel mode",
		testHandler(form.Form{},
			mp3UploadRequest(map[string]string{
				"format":           ".mp3",
				"mp3-channel-mode": "invalid-channel-mode",
			}),
			http.StatusBadRequest),
	)
	t.Run("mp3 invalid vbr quality",
		testHandler(form.Form{},
			mp3UploadRequest(map[string]string{
				"format":            ".mp3",
				"mp3-channel-mode":  "1",
				"mp3-bit-rate-mode": "VBR",
				"mp3-vbr-quality":   "",
			}),
			http.StatusBadRequest),
	)
	t.Run("mp3 invalid quality flag",
		testHandler(form.Form{},
			mp3UploadRequest(map[string]string{
				"format":            ".mp3",
				"mp3-channel-mode":  "1",
				"mp3-bit-rate-mode": "VBR",
				"mp3-vbr-quality":   "1",
				"mp3-use-quality":   "non-bool",
			}),
			http.StatusBadRequest),
	)
	t.Run("mp3 invalid quality flag",
		testHandler(form.Form{},
			mp3UploadRequest(map[string]string{
				"format":            ".mp3",
				"mp3-channel-mode":  "1",
				"mp3-bit-rate-mode": "VBR",
				"mp3-vbr-quality":   "1",
				"mp3-use-quality":   "true",
				"mp3-quality":       "",
			}),
			http.StatusBadRequest),
	)
	t.Run("mp3 invalid VBR bit rate",
		testHandler(form.Form{},
			mp3UploadRequest(map[string]string{
				"format":            ".mp3",
				"mp3-channel-mode":  "1",
				"mp3-bit-rate-mode": "VBR",
				"mp3-bit-rate":      "",
			}),
			http.StatusBadRequest),
	)
	t.Run("mp3 invalid ABR bit rate",
		testHandler(form.Form{},
			mp3UploadRequest(map[string]string{
				"format":            ".mp3",
				"mp3-channel-mode":  "1",
				"mp3-bit-rate-mode": "ABR",
				"mp3-bit-rate":      "",
			}),
			http.StatusBadRequest),
	)
	t.Run("mp3 invalid CBR bit rate",
		testHandler(form.Form{},
			mp3UploadRequest(map[string]string{
				"format":            ".mp3",
				"mp3-channel-mode":  "1",
				"mp3-bit-rate-mode": "CBR",
				"mp3-bit-rate":      "",
			}),
			http.StatusBadRequest),
	)
}
