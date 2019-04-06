package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pipelined/mp3"
	"github.com/pipelined/phono/convert"
	"github.com/pipelined/signal"
	"github.com/rs/xid"
)

const (
	wavMaxSize = 10 * 1024 * 1024
	mp3MaxSize = 1 * 1024 * 1024
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
	VBR           mp3.BitRateMode
	ABR           mp3.BitRateMode
	CBR           mp3.BitRateMode
	BitRateModes  map[mp3.BitRateMode]string
	ChannelModes  map[mp3.ChannelMode]string
	DefineQuality bool
}

var (
	indexTemplate = template.Must(template.ParseFiles("web/index.tmpl"))

	convertFormData = convertForm{
		Accept: fmt.Sprintf(".%s, .%s", convert.WavFormat, convert.Mp3Format),
		OutFormats: []convert.Format{
			convert.WavFormat,
			convert.Mp3Format,
		},
		WavOptions: wavOptions{
			BitDepths: convert.Supported.WavBitDepths,
		},
		Mp3Options: mp3Options{
			VBR:          mp3.VBR,
			ABR:          mp3.ABR,
			CBR:          mp3.CBR,
			BitRateModes: convert.Supported.Mp3BitRateModes,
			ChannelModes: convert.Supported.Mp3ChannelModes,
		},
	}
)

func parseConfig(r *http.Request) (convert.OutputConfig, error) {
	f := convert.Format(r.FormValue("format"))
	switch f {
	case convert.WavFormat:
		return parseWavConfig(r)
	case convert.Mp3Format:
		return parseMp3Config(r)
	default:
		return nil, fmt.Errorf("Unsupported format: %v", f)
	}
}

func parseWavConfig(r *http.Request) (convert.WavConfig, error) {
	// try to get bit depth
	bitDepth, err := parseIntValue(r, "wav-bit-depth", "bit depth")
	if err != nil {
		return convert.WavConfig{}, err
	}

	return convert.WavConfig{BitDepth: signal.BitDepth(bitDepth)}, nil
}

func parseMp3Config(r *http.Request) (convert.Mp3Config, error) {
	// try to get bit rate mode
	bitRateMode, err := parseIntValue(r, "mp3-bit-rate-mode", "bit rate mode")
	if err != nil {
		return convert.Mp3Config{}, err
	}

	// try to get channel mode
	channelMode, err := parseIntValue(r, "mp3-channel-mode", "channel mode")
	if err != nil {
		return convert.Mp3Config{}, err
	}

	if mp3.BitRateMode(bitRateMode) == mp3.VBR {
		// try to get vbr quality
		vbrQuality, err := parseIntValue(r, "mp3-vbr-quality", "vbr quality")
		if err != nil {
			return convert.Mp3Config{}, err
		}
		return convert.Mp3Config{
			BitRateMode: mp3.VBR,
			ChannelMode: mp3.ChannelMode(channelMode),
			VBRQuality:  mp3.VBRQuality(vbrQuality),
		}, nil
	}

	// try to get bitrate
	bitRate, err := parseIntValue(r, "mp3-bit-rate", "bit rate")
	if err != nil {
		return convert.Mp3Config{}, err
	}
	return convert.Mp3Config{
		BitRateMode: mp3.BitRateMode(bitRateMode),
		ChannelMode: mp3.ChannelMode(channelMode),
		BitRate:     bitRate,
	}, nil
}

// parseIntValue parses value of key provided in the html form.
// Returns error if value is not provided or cannot be parsed as int.
func parseIntValue(r *http.Request, key, name string) (int, error) {
	str := r.FormValue(key)
	if str == "" {
		return 0, fmt.Errorf("Please provide %s", name)
	}

	val, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("Failed parsing %s %s: %v", name, str, err)
	}
	return val, nil
}

// convertHandler converts form files to the format provided y form.
func convertHandler(indexTemplate *template.Template, maxSizes map[convert.Format]int64, tmpPath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			indexTemplate.Execute(w, &convertFormData)
		case http.MethodPost:
			// extract input format from the path
			inFormat := convert.Format(path.Base(r.URL.Path))
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
			formFile, handler, err := r.FormFile("input-file")
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid file: %v", err), http.StatusBadRequest)
				return
			}
			defer formFile.Close()

			// parse output config
			outConfig, err := parseConfig(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// create temp file
			tmpFileName := tmpFileName(tmpPath)
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
	dir, err := ioutil.TempDir(".", "phono")
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to create temp folder: %v", err))
	}
	defer os.RemoveAll(dir) // clean up

	// max sizes for different input formats.
	maxSizes := map[convert.Format]int64{
		convert.WavFormat: wavMaxSize,
		convert.Mp3Format: mp3MaxSize,
	}

	// setting router rule
	http.Handle("/", convertHandler(indexTemplate, maxSizes, dir))
	err = http.ListenAndServe(":8080", nil) // setting listening port
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
