package pipes

import (
	"context"
	"fmt"

	"pipelined.dev/pipe"
)

// Encode using Pump as the source and Sinks as destination.
func Encode(ctx context.Context, bufferSize int, pump pipe.SourceAllocatorFunc, sink pipe.SinkAllocatorFunc) error {
	// build encode pipe
	l, err := pipe.Routing{
		Source: pump,
		Sink:   sink,
	}.Line(bufferSize)
	if err != nil {
		return fmt.Errorf("failed to build pipe: %w", err)
	}

	// run conversion
	err = pipe.New(ctx, pipe.WithLines(l)).Wait()
	if err != nil {
		return fmt.Errorf("failed to execute pipe: %w", err)
	}
	return nil
}
