package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var (
	indexTemplate = template.Must(template.ParseFiles("web/index.tmpl"))

	convertForm = &ConvertForm{
		Formats: map[string]string{
			"wav": "wav",
			"mp3": "mp3",
		},
	}
)

const (
	maxInputSize = 2 * 1024 * 1024
	tmpPath      = "tmp"
)

// ConvertForm provides a form for a user to define conversion parameters.
type ConvertForm struct {
	Formats map[string]string
}

func convertHandler(indexTemplate *template.Template, maxSize int64, path string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			indexTemplate.Execute(w, convertForm)
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
				http.Error(w, fmt.Sprintf("Invalid file %v", err), http.StatusBadRequest)
				return
			}
			defer file.Close()

			inFormat := filepath.Ext(handler.Filename)
			fmt.Printf("inFormat: %s\n", inFormat)

			outFormat := r.FormValue("format")
			switch outFormat {
			case "wav":
			case "mp3":
			default:
				http.Error(w, fmt.Sprintf("Invalid output file format: %v", outFormat), http.StatusBadRequest)
				return
			}
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
