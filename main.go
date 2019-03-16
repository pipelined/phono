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

// convertForm provides a form for a user to define conversion parameters.
type convertForm struct {
	Accept     string
	OutFormats []format
	WavOptions wavOptions
}

// format is a file extension.
type format string

// WavOptions is a struct of wav options that are available for conversion.
type wavOptions struct {
	BitDepths map[signal.BitDepth]string
}

// wavConfig is the configuration needed for wav output.
type wavConfig struct {
	signal.BitDepth
}

type config interface {
	Sink(destination) pipe.Sink
}

func (f format) config(r *http.Request) (config, error) {
	switch f {
	case WavFormat:
		return parseWav(r)
	case Mp3Format:
		panic("Not implemented")
	default:
		return nil, fmt.Errorf("Unsupported format: %v", f)
	}
}

func (f format) pump(s source) (pipe.Pump, error) {
	switch f {
	case WavFormat:
		return wav.NewPump(s), nil
	case Mp3Format:
		return mp3.NewPump(s), nil
	default:
		return nil, fmt.Errorf("Unsupported format: %v", f)
	}
}

func parseWav(r *http.Request) (wavConfig, error) {
	// check if bit depth is provided
	bitDepthString := r.FormValue("bit-depth")
	if bitDepthString == "" {
		return wavConfig{}, fmt.Errorf("Please provide bit depth")
	}

	// check if bit depth could be parsed
	bitDepth, err := strconv.Atoi(bitDepthString)
	if err != nil {
		return wavConfig{}, fmt.Errorf("Failed parsing bit depth value %s: %v", bitDepthString, err)
	}

	// check if bit depth is supported
	if _, ok := wavBitDepths[signal.BitDepth(bitDepth)]; !ok {
		return wavConfig{}, fmt.Errorf("Bit depth %v is not supported", bitDepthString)
	}
	return wavConfig{BitDepth: signal.BitDepth(bitDepth)}, nil
}

func (c wavConfig) Sink(d destination) pipe.Sink {
	return wav.NewSink(d, c.BitDepth)
}

var (
	indexTemplate = template.Must(template.ParseFiles("web/index.tmpl"))

	convertFormData = &convertForm{
		Accept: fmt.Sprintf("%s, %s", WavFormat, Mp3Format),
		OutFormats: []format{
			WavFormat,
			Mp3Format,
		},
		WavOptions: wavOptions{
			BitDepths: wavBitDepths,
		},
	}

	wavBitDepths = map[signal.BitDepth]string{
		signal.BitDepth8:  "8 bit",
		signal.BitDepth16: "16 bits",
		signal.BitDepth24: "24 bits",
		signal.BitDepth32: "32 bits",
	}
)

const (
	maxInputSize = 2 * 1024 * 1024
	tmpPath      = "tmp"

	// WavFormat represents .wav files
	WavFormat = format(".wav")
	// Mp3Format represents .mp3 files
	Mp3Format = format(".mp3")
)

type source interface {
	io.Reader
	io.Seeker
	io.Closer
}

type destination interface {
	io.Writer
	io.Seeker
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
			formFile, handler, err := r.FormFile("convertfile")
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid file: %v", err), http.StatusBadRequest)
				return
			}
			defer formFile.Close()
			inFormat := format(filepath.Ext(handler.Filename))

			// create temp file
			outFormat := format(r.FormValue("format"))
			outConfig, err := outFormat.config(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			tmpFileName := tmpFileName(path)
			tmpFile, err := os.Create(tmpFileName)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error creating temp file: %v", err), http.StatusInternalServerError)
				return
			}
			defer cleanUp(tmpFile)

			// convert file using temp file
			err = convert(formFile, tmpFile, inFormat, outConfig)
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
func outFileName(name string, oldExt, newExt format) string {
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

func convert(s source, d destination, sourceFormat format, destinationConfig config) error {
	// create pump for input format
	pump, err := sourceFormat.pump(s)
	if err != nil {
		return fmt.Errorf("Unsupported input format: %s", sourceFormat)
	}
	// create sink for output format
	sink := destinationConfig.Sink(d)

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
