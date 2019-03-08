package main

import (
	"html/template"
	"log"
	"net/http"
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

// ConvertForm provides a form for a user to define conversion parameters.
type ConvertForm struct {
	Formats map[string]string
}

func indexHandler(indexTemplate *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			indexTemplate.Execute(w, convertForm)
		case http.MethodPost:
			r.ParseForm()
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func main() {
	// setting router rule
	http.Handle("/", indexHandler(indexTemplate))
	err := http.ListenAndServe(":8080", nil) // setting listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
