package convert

import (
	"fmt"

	"github.com/pipelined/pipe"
)

// Convert using pump as the source and SinkBuilder as destination.
func Convert(pump pipe.Pump, sinks ...pipe.Sink) error {
	// build convert pipe
	convert, err := pipe.New(1024, pipe.WithPump(pump), pipe.WithSinks(sinks...))
	if err != nil {
		return fmt.Errorf("Failed to build pipe: %v", err)
	}

	// run conversion
	err = pipe.Wait(convert.Run())
	if err != nil {
		return fmt.Errorf("Failed to execute pipe: %v", err)
	}
	return nil
}
