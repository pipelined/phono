package encode

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"strconv"

	"github.com/pipelined/phono/encode/form"
	"github.com/pipelined/phono/pipes"
	"pipelined.dev/audio/fileformat"
)

var (
	errInputFormat  = errors.New("unsupported input format")
	errOutputFormat = errors.New("unsupported output format")
)

// Handler form files to the format provided by form.
// Process request steps:
//	1. Retrieve input format from URL
//	2. Use http.MaxBytesReader to avoid memory abuse
//	3. Parse output configuration
//	4. Create temp file
//	5. Run conversion
//	6. Send result file
func Handler(f form.Form, bufferSize int, tempDir string) http.Handler {
	formData := f.Data()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, err := w.Write(formData)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case http.MethodPost:
			inputFormat, ok := fileformat.FormatByPath(r.URL.Path)
			if !ok {
				http.Error(w, errInputFormat.Error(), http.StatusBadRequest)
				return
			}
			// get max size for the format
			maxSize := f.InputMaxSize(inputFormat)
			if maxSize > 0 {
				// check if limit is defined
				if maxSize > 0 {
					r.Body = http.MaxBytesReader(w, r.Body, maxSize)
				}
			}
			// check max size
			if err := r.ParseMultipartForm(maxSize); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			f, _, err := r.FormFile(form.FormFileKey)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer f.Close()

			// parse sink and validate parameters
			sink, ext, err := form.ParseForm(r.MultipartForm.Value)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			outputFormat, ok := fileformat.FormatByPath(ext)
			if !ok {
				http.Error(w, errOutputFormat.Error(), http.StatusBadRequest)
			}

			// create temp file
			tempFile, err := ioutil.TempFile(tempDir, "")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer cleanUp(tempFile)

			// encode file using temp file
			if err = pipes.Encode(r.Context(), bufferSize, inputFormat.Source(f), sink(tempFile)); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			// reset temp file
			_, err = tempFile.Seek(0, 0)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to reset temp file: %v", err), http.StatusInternalServerError)
				return
			}
			// get temp file stats for headers
			stat, err := tempFile.Stat()
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to get file stats: %v", err), http.StatusInternalServerError)
				return
			}
			fileSize := strconv.FormatInt(stat.Size(), 10)
			//Send the headers
			w.Header().Set("Content-Disposition", "attachment; filename="+outFileName("result", 1, outputFormat.DefaultExtension()))
			w.Header().Set("Content-Type", mime.TypeByExtension(outputFormat.DefaultExtension()))
			w.Header().Set("Content-Length", fileSize)
			_, err = io.Copy(w, tempFile) // send file to a client
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to transfer file: %v", err), http.StatusInternalServerError)
			}
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
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
