package handler_test

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

	"golang.org/x/net/html"

	"github.com/pipelined/phono/form"
	"github.com/pipelined/phono/handler"
	"github.com/stretchr/testify/assert"
)

func parseURL(raw string) (result *url.URL) {
	result, _ = url.Parse(raw)
	return
}

// Creates a new file upload http request with optional extra params. Any error causes panic.
func newFileUploadRequest(uri string, params map[string]string, fileKey, filePath string) *http.Request {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fileKey, filepath.Base(filePath))
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

func TestEncode(t *testing.T) {
	tests := []struct {
		form           form.Encode
		r              *http.Request
		expectedStatus int
		tempDir        string
	}{
		{
			r: &http.Request{
				Method: http.MethodPut,
				URL:    parseURL("test/.test"),
			},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			r: &http.Request{
				Method: http.MethodPost,
				URL:    parseURL("test/.test"),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			r: &http.Request{
				Method: http.MethodPost,
				URL:    parseURL("test/.wav"),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			r: newFileUploadRequest(
				"test/.wav",
				nil,
				form.FileKey,
				"../_testdata/sample.wav"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			form: form.Encode{WavMaxSize: 10},
			r: newFileUploadRequest(
				"test/.wav",
				map[string]string{
					"format":        ".wav",
					"wav-bit-depth": "16",
				},
				form.FileKey,
				"../_testdata/sample.wav"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			r: newFileUploadRequest(
				"test/.wav",
				map[string]string{
					"format":        ".wav",
					"wav-bit-depth": "16",
				},
				"wrong-file-key",
				"../_testdata/sample.wav"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			r: newFileUploadRequest(
				"test/.wav",
				map[string]string{
					"format":        ".wav",
					"wav-bit-depth": "16",
				},
				form.FileKey,
				"../_testdata/not-media"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			tempDir: "non-existent-directory",
			r: newFileUploadRequest(
				"test/.wav",
				map[string]string{
					"format":        ".wav",
					"wav-bit-depth": "16",
				},
				form.FileKey,
				"../_testdata/sample.wav"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			r: newFileUploadRequest(
				"test/.wav",
				map[string]string{
					"format":        ".wav",
					"wav-bit-depth": "16",
				},
				form.FileKey,
				"../_testdata/sample.wav"),
			expectedStatus: http.StatusOK,
		},
	}
	bufferSize := 1024
	for _, test := range tests {
		h := handler.Encode(test.form, bufferSize, test.tempDir)
		assert.NotNil(t, h)

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, test.r)
		assert.Equal(t, test.expectedStatus, rr.Code)
	}
}

func TestEncodeForm(t *testing.T) {
	h := handler.Encode(form.Encode{}, 1024, "")
	assert.NotNil(t, h)

	r, _ := http.NewRequest(http.MethodGet, "test/", nil)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, r)
	assert.Equal(t, http.StatusOK, rr.Code)
	_, err := html.Parse(rr.Body)
	assert.Nil(t, err)
}
