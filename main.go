package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pipelined/mp3"
	"github.com/pipelined/pipe"
	"github.com/pipelined/signal"
	"github.com/pipelined/wav"
	"github.com/rs/xid"
)

var (
	indexTemplate = template.Must(template.ParseFiles("web/index.tmpl"))

	convertFormData = &ConvertForm{
		Accept: fmt.Sprintf("%s, %s", WavFormat, Mp3Format),
		OutFormats: []Format{
			WavFormat,
			Mp3Format,
		},
	}
)

const (
	maxInputSize = 2 * 1024 * 1024
	tmpPath      = "tmp"

	// WavFormat represents .wav files
	WavFormat = Format(".wav")
	// Mp3Format represents .mp3 files
	Mp3Format = Format(".mp3")
)

// ConvertForm provides a form for a user to define conversion parameters.
type ConvertForm struct {
	Accept     string
	OutFormats []Format
}

// Format is a file extension.
type Format string

type source interface {
	io.Reader
	io.Seeker
	io.Closer
}

// convertHandler converts form files to the format provided y form.
func convertHandler(indexTemplate *template.Template, maxSize int64, path string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			indexTemplate.Execute(w, convertFormData)
		case http.MethodPost:
			// check max size
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)
			if err := r.ParseMultipartForm(maxSize); err != nil {
				http.Error(w, "File too big", http.StatusBadRequest)
				return
			}
			// obtain file handler
			file, handler, err := r.FormFile("convertfile")
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid file: %v", err), http.StatusBadRequest)
				return
			}
			defer file.Close()
			inFormat := Format(filepath.Ext(handler.Filename))

			// create temp file
			outFormat := Format(r.FormValue("format"))
			tmpFileName := tmpFileName(path)
			tmpFile, err := os.Create(tmpFileName)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error creating temp file: %v", err), http.StatusInternalServerError)
				return
			}
			defer cleanUp(tmpFile)

			// convert file using temp file
			err = convert(file, tmpFile, inFormat, outFormat)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
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
			w.Header().Set("Content-Disposition", "attachment; filename="+outFileName(handler.Filename, inFormat, outFormat))
			w.Header().Set("Content-Type", mime.TypeByExtension(string(outFormat)))
			w.Header().Set("Content-Length", fileSize)
			io.Copy(w, tmpFile) // send file to a client
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func main() {
	// setting router rule
	http.Handle("/", convertHandler(indexTemplate, maxInputSize, tmpPath))
	err := http.ListenAndServe(":8080", nil) // setting listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// tmpFileName returns temporary file name. It uses xid library to generate names on the fly.
func tmpFileName(path string) string {
	return filepath.Join(path, xid.New().String())
}

// outFileName return output file name. It replaces input format extension with output.
func outFileName(name string, oldExt, newExt Format) string {
	return strings.Replace(strings.ToLower(name), string(oldExt), string(newExt), 1)
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

func convert(s source, destination *os.File, inFormat, outFormat Format) error {
	// create pump for input format
	var pump pipe.Pump
	switch inFormat {
	case WavFormat:
		pump = wav.NewPump(s)
	case Mp3Format:
		pump = mp3.NewPump(s)
	default:
		return fmt.Errorf("Invalid input file format: %v", inFormat)
	}

	// create sink for output format
	var sink pipe.Sink
	switch outFormat {
	case ".wav":
		sink = wav.NewSink(destination, signal.BitDepth24)
	case ".mp3":
		sink = mp3.NewSink(destination, 192, 10)
	default:
		return fmt.Errorf("Invalid output file format: %v", outFormat)
	}

	// build convert pipe
	convert, err := pipe.New(1024, pipe.WithPump(pump), pipe.WithSinks(sink))
	if err != nil {
		return fmt.Errorf("Failed to build pipe: %v", err)
	}

	// run conversion
	err = pipe.Wait(convert.Run())
	if err != nil {
		return fmt.Errorf("Failed to execute pipe: %v", err)
	}
	return nil
}
