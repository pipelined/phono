package controller

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"strconv"

	"github.com/pipelined/mp3"
	"github.com/pipelined/phono/convert"
	"github.com/pipelined/pipe"
	"github.com/pipelined/wav"
)

type ConvertForm interface {
	Data() []byte
	FileKey() string
	ParseExtension(*http.Request) string
	ParsePump(fileName string) (pipe.Pump, error)
	ParseOutput(*http.Request) (pipe.Sink, string, error)
}

// Convert converts form files to the format provided y form.
// To limit maximum input file size, pass map of extensions with limits.
// Process request steps:
//	1. Retrieve input format from URL
//	2. Use http.MaxBytesReader to avoid memory abuse
//	3. Parse output configuration
//	4. Create temp file
//	5. Run conversion
//	6. Send result file
func Convert(convertForm ConvertForm, limits map[string]int64, tempDir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, err := w.Write(convertForm.Data())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case http.MethodPost:
			// extract input format from the path
			inExt := convertForm.ParseExtension(r)
			// get max size for the format
			if maxSize, ok := limits[inExt]; ok {
				r.Body = http.MaxBytesReader(w, r.Body, maxSize)
				// check max size
				if err := r.ParseMultipartForm(maxSize); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			} else {
				http.Error(w, fmt.Sprintf("Format %s not supported", inExt), http.StatusBadRequest)
				return
			}

			// get form file
			f, handler, err := r.FormFile(convertForm.FileKey())
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid file: %v", err), http.StatusBadRequest)
				return
			}
			defer f.Close()

			// obtain file handler
			pump, err := convertForm.ParsePump(handler.Filename)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if err = assignFormFile(f, pump); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// parse output config
			sink, outExt, err := convertForm.ParseOutput(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			tmpFile, err := createTempFile(tempDir, sink)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer cleanUp(tmpFile)

			// convert file using temp file
			err = convert.Convert(pump, sink)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// reset temp file
			_, err = tmpFile.Seek(0, 0)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to reset temp file: %v", err), http.StatusInternalServerError)
				return
			}
			// get temp file stats for headers
			stat, err := tmpFile.Stat()
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to get file stats: %v", err), http.StatusInternalServerError)
				return
			}
			fileSize := strconv.FormatInt(stat.Size(), 10)
			//Send the headers
			w.Header().Set("Content-Disposition", "attachment; filename="+outFileName("result", 1, outExt))
			w.Header().Set("Content-Type", mime.TypeByExtension(outExt))
			w.Header().Set("Content-Length", fileSize)
			_, err = io.Copy(w, tmpFile) // send file to a client
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to transfer file: %v", err), http.StatusInternalServerError)
			}
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func assignFormFile(r io.ReadSeeker, p pipe.Pump) (err error) {
	switch v := p.(type) {
	case *wav.Pump:
		v.ReadSeeker = r
	case *mp3.Pump:
		v.Reader = r
	default:
		err = fmt.Errorf("%T sink is not supported", v)
	}
	return
}

func createTempFile(dir string, s pipe.Sink) (f *os.File, err error) {
	switch v := s.(type) {
	case *wav.Sink:
		f, err = ioutil.TempFile(dir, "")
		v.WriteSeeker = f
		return
	case *mp3.Sink:
		f, err = ioutil.TempFile(dir, "")
		v.Writer = f
		return
	default:
		err = fmt.Errorf("%T sink is not supported", v)
		return
	}
}

// outFileName return output file name. It replaces input format extension with output.
func outFileName(prefix string, idx int, ext string) string {
	return fmt.Sprintf("%v_%d%v", prefix, idx, ext)
}

// cleanUp removes temporary file and handles all errors on the way.
func cleanUp(f *os.File) {
	err := f.Close()
	if err != nil {
		log.Printf("Failed to close temp file")
	}
	err = os.Remove(f.Name())
	if err != nil {
		log.Printf("Failed to delete temp file")
	}
}
