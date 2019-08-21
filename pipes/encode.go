package pipes

import (
	"context"
	"fmt"

	"github.com/pipelined/pipe"
)

// Encode using Pump as the source and Sinks as destination.
func Encode(ctx context.Context, bufferSize int, pump pipe.Pump, sinks ...pipe.Sink) error {
	// build encode pipe
	l, err := pipe.Line(
		&pipe.Pipe{
			Pump:  pump,
			Sinks: sinks,
		},
	)
	if err != nil {
		return fmt.Errorf("Failed to build pipe: %v", err)
	}

	// run conversion
	err = pipe.Wait(l.Run(ctx, bufferSize))
	if err != nil {
		return fmt.Errorf("Failed to execute pipe: %v", err)
	}
	return pipe.Wait(l.Close())
}
