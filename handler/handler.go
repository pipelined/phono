package handler

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pipelined/mp3"
	"github.com/pipelined/phono/convert"
	"github.com/pipelined/phono/template"
	"github.com/pipelined/wav"
	"github.com/rs/xid"
)

// Convert converts form files to the format provided y form.
func Convert(convertForm template.ConvertForm, maxSizes map[string]int64, tmpPath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, err := w.Write(convertForm.Data())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case http.MethodPost:
			// extract input format from the path
			inFormat := convertForm.Format(r)
			// get max size for the format
			if maxSize, ok := maxSizes[inFormat]; ok {
				r.Body = http.MaxBytesReader(w, r.Body, maxSize)
				// check max size
				if err := r.ParseMultipartForm(maxSize); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			} else {
				http.Error(w, fmt.Sprintf("Format %s not supported", inFormat), http.StatusBadRequest)
				return
			}

			// obtain file handler
			pump, closer, err := convertForm.Pump(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer closer.Close()

			// parse output config
			outConfig, err := convertForm.Parse(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// create temp file
			tmpFile, err := createTmpFile(outConfig, tmpPath)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error creating temp file: %v", err), http.StatusInternalServerError)
				return
			}
			defer cleanUp(tmpFile)

			// convert file using temp file
			err = convert.Convert(pump, outConfig)
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
			w.Header().Set("Content-Disposition", "attachment; filename="+outFileName("test", inFormat, ".wav"))
			w.Header().Set("Content-Type", mime.TypeByExtension(".wav"))
			w.Header().Set("Content-Length", fileSize)
			io.Copy(w, tmpFile) // send file to a client
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func createTmpFile(output convert.SinkBuilder, path string) (*os.File, error) {
	f, err := os.Create(filepath.Join(path, xid.New().String()))
	if err != nil {
		return nil, err
	}
	switch config := output.(type) {
	case *mp3.SinkBuilder:
		config.Writer = f
	case *wav.SinkBuilder:
		config.WriteSeeker = f
	}
	return f, nil
}

// outFileName return output file name. It replaces input format extension with output.
func outFileName(name string, oldExt, newExt string) string {
	return strings.Replace(strings.ToLower(name), oldExt, newExt, 1)
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
