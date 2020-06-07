package util

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func GracefulContext(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-ctx.Done():
		case <-c:
		}
	}()
	return ctx
}
