package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/pipelined/phono/controller"
	"github.com/pipelined/phono/input/form"

	"github.com/spf13/cobra"
)

const (
	wavMaxSize = 10 * 1024 * 1024
	mp3MaxSize = 1 * 1024 * 1024
	bufferSize = 1024
)

var (
	convertCmd = &cobra.Command{
		Use:   "convert",
		Short: "Convert audio files",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	convertHTTPPort int
	convertHTTPCmd  = &cobra.Command{
		Use:   "http",
		Short: "Spin up the http service to convert files",
		Run: func(cmd *cobra.Command, args []string) {
			serve(convertHTTPPort)
		},
	}
)

func init() {
	convertHTTPCmd.Flags().IntVar(&convertHTTPPort, "port", 8080, "Start convert http handler")
	convertCmd.AddCommand(convertHTTPCmd)
	root.AddCommand(convertCmd)
}

func serve(port int) {
	// temporary directory
	dir, err := ioutil.TempDir("", "phono")
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to create temp folder: %v", err))
	}

	interrupt := make(chan struct{})
	server := http.Server{Addr: fmt.Sprintf(":%d", port)}
	go func() {
		sigint := make(chan os.Signal, 1)

		// interrupt and sigterm signal
		signal.Notify(sigint, os.Interrupt)
		signal.Notify(sigint, syscall.SIGTERM)

		// block until signal received
		<-sigint

		// interrupt signal received, shut down
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown error: %v", err)
		}
		close(interrupt)
	}()

	// max sizes for different input formats.

	// setting router rule
	http.Handle("/", controller.Convert(form.Convert{WavMaxSize: wavMaxSize, Mp3MaxSize: mp3MaxSize}, bufferSize, dir))
	log.Printf("phono convert at: http://localhost%s\n", server.Addr)
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
