package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"pipelined.dev/phono/encode"
	"pipelined.dev/phono/userinput"
)

var (
	encodeHTTP = struct {
		port       int
		tempDir    string
		bufferSize int
	}{}
	encodeHTTPCmd = &cobra.Command{
		Use:   "http",
		Short: "Spin up the http service to encode files",
		Run: func(cmd *cobra.Command, args []string) {
			serve(encodeHTTP.port, encodeHTTP.tempDir, encodeHTTP.bufferSize)
		},
	}
)

func init() {
	encodeCmd.AddCommand(encodeHTTPCmd)
	encodeHTTPCmd.Flags().IntVar(&encodeHTTP.port, "port", 8080, "port to use")
	encodeHTTPCmd.Flags().StringVar(&encodeHTTP.tempDir, "tempdir", "", "directory for temp files. defaults to os.TempDir if empty")
	encodeHTTPCmd.Flags().IntVar(&encodeHTTP.bufferSize, "buffersize", 1024, "buffer size")
}

func serve(port int, tempDir string, bufferSize int) {
	// temporary directory
	dir, err := ioutil.TempDir(tempDir, "phono")
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to create temp folder: %v", err))
	}

	// setting router rule
	mux := http.NewServeMux()
	mux.Handle("/", encode.Handler(userinput.NewEncodeForm(userinput.Limits{}), bufferSize, dir))
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	interrupted := onInterrupt(func() {
		// interrupt signal received, shut down
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown error: %v", err)
		}
	})

	log.Printf("phono encode at: http://localhost%s\n", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("HTTP server ListenAndServe error: %v", err)
	}

	// block until shutdown executed
	<-interrupted

	// clean up
	err = os.RemoveAll(dir)
	if err != nil {
		log.Printf("Clean up error: %v", err)
	}
}
