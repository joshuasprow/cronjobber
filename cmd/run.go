package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joshuasprow/cronjobber/pkg"
	"github.com/joshuasprow/cronjobber/pkg/k8s"
	"github.com/joshuasprow/cronjobber/pkg/templates"
)

func Run(log *slog.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientset, err := k8s.NewClientset("")
	if err != nil {
		return fmt.Errorf("new clientset: %w", err)
	}

	tmpl, err := templates.New()
	if err != nil {
		return fmt.Errorf("new templates: %w", err)
	}

	server := pkg.NewServer(log, clientset, tmpl)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel()
	}()

	if err := server.Run(ctx); err != nil {
		return fmt.Errorf("run server: %w", err)
	}

	return nil
}
