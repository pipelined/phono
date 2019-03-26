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

	"github.com/pipelined/convert"
	"github.com/pipelined/mp3"
	"github.com/pipelined/signal"
	"github.com/rs/xid"
)

// convertForm provides a form for a user to define conversion parameters.
type convertForm struct {
	Accept     string
	OutFormats []convert.Format
	WavOptions wavOptions
	Mp3Options mp3Options
}

// WavOptions is a struct of wav options that are available for conversion.
type wavOptions struct {
	BitDepths map[signal.BitDepth]string
}

type mp3Options struct {
	BitRateModes  map[mp3.BitRateMode]string
	ChannelModes  map[mp3.ChannelMode]string
	DefineQuality bool
	Qualities     map[mp3.Quality]string
	VBRQualities  map[mp3.VBRQuality]string
}

var (
	indexTemplate = template.Must(template.ParseFiles("web/index.tmpl"))

	convertFormData = convertForm{
		Accept: fmt.Sprintf("%s, %s", convert.WavFormat, convert.Mp3Format),
		OutFormats: []convert.Format{
			convert.WavFormat,
			convert.Mp3Format,
		},
		WavOptions: wavOptions{
			BitDepths: convert.Supported.WavBitDepths,
		},
		Mp3Options: mp3Options{
			BitRateModes: convert.Supported.Mp3BitRateModes,
			ChannelModes: convert.Supported.Mp3ChannelModes,
			VBRQualities: convert.Supported.Mp3VBRQualities,
			Qualities:    convert.Supported.Mp3Qualities,
		},
	}
)

func parseConfig(r *http.Request) (convert.OutputConfig, error) {
	f := convert.Format(r.FormValue("format"))
	switch f {
	case convert.WavFormat:
		return parseWavConfig(r)
	case convert.Mp3Format:
		return convert.Mp3Config{}, nil
	default:
		return nil, fmt.Errorf("Unsupported format: %v", f)
	}
}

func parseWavConfig(r *http.Request) (convert.WavConfig, error) {
	// check if bit depth is provided
	bitDepthString := r.FormValue("wav-bit-depth")
	if bitDepthString == "" {
		return convert.WavConfig{}, fmt.Errorf("Please provide bit depth")
	}

	// check if bit depth could be parsed
	bitDepth, err := strconv.Atoi(bitDepthString)
	if err != nil {
		return convert.WavConfig{}, fmt.Errorf("Failed parsing bit depth value %s: %v", bitDepthString, err)
	}

	// check if bit depth is supported
	if _, ok := convert.Supported.WavBitDepths[signal.BitDepth(bitDepth)]; !ok {
		return convert.WavConfig{}, fmt.Errorf("Bit depth %v is not supported", bitDepthString)
	}
	return convert.WavConfig{BitDepth: signal.BitDepth(bitDepth)}, nil
}

const (
	maxInputSize = 2 * 1024 * 1024
)

// convertHandler converts form files to the format provided y form.
func convertHandler(indexTemplate *template.Template, maxSize int64, path string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			indexTemplate.Execute(w, &convertFormData)
		case http.MethodPost:
			// check max size
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)
			if err := r.ParseMultipartForm(maxSize); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// obtain file handler
			formFile, handler, err := r.FormFile("convertfile")
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid file: %v", err), http.StatusBadRequest)
				return
			}
			defer formFile.Close()
			inFormat := convert.Format(filepath.Ext(handler.Filename))

			// parse output config
			outConfig, err := parseConfig(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// create temp file
			tmpFileName := tmpFileName(path)
			tmpFile, err := os.Create(tmpFileName)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error creating temp file: %v", err), http.StatusInternalServerError)
				return
			}
			defer cleanUp(tmpFile)

			// convert file using temp file
			err = convert.Convert(formFile, tmpFile, inFormat, outConfig)
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
			w.Header().Set("Content-Disposition", "attachment; filename="+outFileName(handler.Filename, inFormat, outConfig.Format()))
			w.Header().Set("Content-Type", mime.TypeByExtension(string(outConfig.Format())))
			w.Header().Set("Content-Length", fileSize)
			io.Copy(w, tmpFile) // send file to a client
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func main() {
	tmpPath := "tmp"
	if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
		err = os.Mkdir(tmpPath, os.ModePerm)
		if err != nil {
			log.Fatal(fmt.Sprintf("Failed to create temp folder: %v", err))
		}
	}

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
func outFileName(name string, oldExt, newExt convert.Format) string {
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
