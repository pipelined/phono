package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/pipelined/phono/handler"

	"github.com/pipelined/phono/convert"
)

const (
	wavMaxSize = 10 * 1024 * 1024
	mp3MaxSize = 1 * 1024 * 1024
)

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
	http.Handle("/", handler.Convert(maxSizes, dir))
	err = http.ListenAndServe(":8080", nil) // setting listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
