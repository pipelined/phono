package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

var (
	loginTemplate *template.Template
)

func index(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		loginTemplate.Execute(w, nil)
	case http.MethodPost:
		r.ParseForm()
		fmt.Println("username:", r.Form["username"])
		fmt.Println("password:", r.Form["password"])
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	loginTemplate, _ = template.ParseFiles("web/index.tmpl")
	// setting router rule
	http.HandleFunc("/", index)
	err := http.ListenAndServe(":8080", nil) // setting listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
