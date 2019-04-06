package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
	// temporary directory
	dir, err := ioutil.TempDir(".", "phono")
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to create temp folder: %v", err))
	}

	interrupt := make(chan struct{})
	var server http.Server
	go func() {
		sigint := make(chan os.Signal, 1)

		// interrupt and sigterm signal
		signal.Notify(sigint, os.Interrupt)
		signal.Notify(sigint, syscall.SIGTERM)

		// block until signal received
		<-sigint

		// interrupt signal received, shut down
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(interrupt)
	}()

	// max sizes for different input formats.
	maxSizes := map[convert.Format]int64{
		convert.WavFormat: wavMaxSize,
		convert.Mp3Format: mp3MaxSize,
	}

	// setting router rule
	http.Handle("/", handler.Convert(template.ConvertForm, maxSizes, dir))
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("HTTP server ListenAndServe error: %v", err)
	}

	// block until shutdown executed
	<-interrupt

	// clean up
	err = os.RemoveAll(dir)
	if err != nil {
		log.Printf("Clean up error: %v", err)
	}
}
