package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Dokhoyan/marketplace/internal/app"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	a, err := app.NewApp(ctx)
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}
	if err := a.Run(ctx); err != nil {
		log.Fatalf("failed to run app: %v", err)
	}
}
