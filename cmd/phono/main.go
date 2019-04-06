package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/pipelined/phono/cmd"
	"github.com/pipelined/phono/handler"
	"github.com/pipelined/phono/template"
	"github.com/spf13/cobra"

	"github.com/pipelined/phono/convert"
)

const (
	wavMaxSize = 10 * 1024 * 1024
	mp3MaxSize = 1 * 1024 * 1024
)

var (
	convertCmd = &cobra.Command{
		Use:   "convert",
		Short: "audio files",
		Run: func(cmd *cobra.Command, args []string) {
			if convertHTTP {
				serve()
			}
		},
	}

	convertHTTP bool
)

func main() {
	convertCmd.Flags().BoolVar(&convertHTTP, "http", false, "start convert http handler")
	cmd.Root.AddCommand(convertCmd)
	cmd.Execute()
}

func serve() {
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
	http.Handle("/", handler.Convert(template.ConvertForm, maxSizes, dir))
	err = http.ListenAndServe(":8080", nil) // setting listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
