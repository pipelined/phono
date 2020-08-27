package encode

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"pipelined.dev/audio/fileformat"
	"pipelined.dev/pipe"
)

type (
	Form interface {
		Bytes() []byte
		Parse(*http.Request) (FormData, error)
	}

	// FormData contains parsed form data.
	FormData struct {
		Input
		Output
	}

	Input struct {
		fileformat.Format
		multipart.File
	}
	Output struct {
		fileformat.Format
		Sink func(io.WriteSeeker) pipe.SinkAllocatorFunc
	}
)

// Handler form files to the format provided by form.
// Process request steps:
//	1. Retrieve userinput format from URL
//	2. Use http.MaxBytesReader to avoid memory abuse
//	3. Parse output configuration
//	4. Create temp file
//	5. Run conversion
//	6. Send result file
func Handler(f Form, bufferSize int, tempDir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, err := w.Write(f.Bytes())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case http.MethodPost:
			formData, err := f.Parse(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer formData.Close()

			// create temp file
			tempFile, err := ioutil.TempFile(tempDir, "")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer cleanUp(tempFile)

			// encode file using temp file
			if err = Run(r.Context(), bufferSize, formData.Input.Source(formData.File), formData.Output.Sink(tempFile)); err != nil {
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
			w.Header().Set("Content-Disposition", "attachment; filename="+outFileName("result", 1, formData.Output.DefaultExtension()))
			w.Header().Set("Content-Type", mime.TypeByExtension(formData.Output.DefaultExtension()))
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

// outFileName return output file name. It replaces userinput format extension with output.
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
