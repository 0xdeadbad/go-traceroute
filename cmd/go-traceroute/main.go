package main

import (
	"context"
	"internal/platform"
	"os"

	"pkg/traceroute"
)

func main() {
	ctx, cancel := platform.SignalNotifyContext(context.Background())

	t, err := traceroute.NewTracer(ctx, os.Args[1])
	if err != nil {
		panic(err)
	}

	err = t.Start()
	if err != nil {
		panic(err)
	}

	cancel()

	<-ctx.Done()
}
