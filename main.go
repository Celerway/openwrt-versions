package main

import (
	"context"
	"flag"
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

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	arch := fs.String("arch", "x86/64", "Architecture for upstream package lists")
	age := fs.String("version", "", "Version for upstream")

	err := fs.Parse(args)
	if err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}
	// check that the version is not empty
	if *age == "" {
		return fmt.Errorf("version cannot be empty")
	}
	fmt.Println("arch: ", *arch)
	fmt.Println("version: ", *age)
	return nil
}
