package encode

import (
	"context"
	"fmt"

	"pipelined.dev/pipe"
)

// Run encoding using Pump as the source and Sinks as destination.
func Run(ctx context.Context, bufferSize int, pump pipe.SourceAllocatorFunc, sink pipe.SinkAllocatorFunc) error {
	// run conversion
	err := pipe.Run(ctx, bufferSize, pipe.Line{
		Source: pump,
		Sink:   sink,
	})
	if err != nil {
		return fmt.Errorf("failed to execute pipe: %w", err)
	}
	return nil
}
