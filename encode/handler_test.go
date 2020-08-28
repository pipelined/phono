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

	"pipelined.dev/phono/encode"
	"pipelined.dev/phono/userinput"
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
	part, err := writer.CreateFormFile(userinput.FormFileKey, filepath.Base(filePath))
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

func notMediaUploadRequest(uri string, params map[string]string) *http.Request {
	return fileUploadRequest(uri, params, "../_testdata/not-media")
}

func TestHandler(t *testing.T) {
	f := userinput.NewEncodeForm(userinput.Limits{})
	bufferSize := 512
	testHandler := func(l encode.Form, r *http.Request, expectedStatus int) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()
			h := encode.Handler(l, bufferSize, "")
			assert.NotNil(t, h)

			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, r)
			assert.Equal(t, expectedStatus, rr.Code)
		}
	}
	t.Run("not allowed method",
		testHandler(f,
			&http.Request{
				Method: http.MethodPut,
				URL:    parseURL("test/.wav"),
			},
			http.StatusMethodNotAllowed),
	)
	t.Run("not allowed userinput format",
		testHandler(f,
			&http.Request{
				Method: http.MethodPost,
				URL:    parseURL("test/.test"),
			},
			http.StatusBadRequest),
	)
	t.Run("wav empty body",
		testHandler(f,
			&http.Request{
				Method: http.MethodPost,
				URL:    parseURL("test/.wav"),
			},
			http.StatusBadRequest),
	)
	t.Run("wav missing bit depth",
		testHandler(f,
			wavUploadRequest(nil),
			http.StatusBadRequest),
	)
	t.Run("wav ok",
		testHandler(f,
			wavUploadRequest(map[string]string{
				"format":        ".wav",
				"wav-bit-depth": "16",
			}),
			http.StatusOK),
	)
}
