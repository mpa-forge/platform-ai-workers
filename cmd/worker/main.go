package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mpa-forge/platform-ai-workers/internal/app"
	"github.com/mpa-forge/platform-ai-workers/internal/config"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] != "run" {
		log.Fatalf("unsupported command %q; expected \"run\"", os.Args[1])
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	worker, err := app.New(cfg)
	if err != nil {
		log.Fatalf("worker bootstrap error: %v", err)
	}

	result, err := worker.Run(ctx)
	if err != nil {
		log.Fatalf("worker run failed: %v", err)
	}

	fmt.Printf("worker finished with result=%s\n", result)
}
