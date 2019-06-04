package pipes

import (
	"context"
	"fmt"

	"github.com/pipelined/pipe"
)

// Encode using Pump as the source and Sinks as destination.
func Encode(ctx context.Context, bufferSize int, pump pipe.Pump, sink ...pipe.Sink) error {
	// build encode pipe
	p, err := pipe.New(bufferSize,
		pipe.WithPump(pump),
		pipe.WithSinks(sink...),
	)
	if err != nil {
		return fmt.Errorf("Failed to build pipe: %v", err)
	}

	// run conversion
	errc := p.Run()
	select {
	case err, ok := <-errc:
		if ok {
			return fmt.Errorf("Failed to execute pipe: %v", err)
		}
		return nil
	case <-ctx.Done():
		return pipe.Wait(p.Close())
	}
}
