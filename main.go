package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	if err := run(ctx, os.Stdout, os.Args[1:]); err != nil {
		fmt.Println("error: ", err)
	}
	os.Exit(0)
}

func run(ctx context.Context, stdout *os.File, args []string) error {
	// Do something
	return nil
}
