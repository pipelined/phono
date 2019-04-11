package convert

import (
	"fmt"

	"github.com/pipelined/pipe"
)

// SinkBuilder builds new Sink. It also validates configuration during Build() call.
type SinkBuilder interface {
	Build() (pipe.Sink, error)
}

// Convert provided source of sourceFormat into destination using destinationConfig.
func Convert(pump pipe.Pump, builder SinkBuilder) error {
	// create sink for output format
	sink, err := builder.Build()
	if err != nil {
		return fmt.Errorf("Provided configuration is not supported")
	}

	// build convert pipe
	convert, err := pipe.New(1024, pipe.WithPump(pump), pipe.WithSinks(sink))
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
