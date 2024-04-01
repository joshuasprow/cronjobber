package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/joshuasprow/cronjobber/pkg/templates"
	"k8s.io/client-go/kubernetes"
)

type Server struct {
	s *http.Server
}

func New(
	log *slog.Logger,
	clientset *kubernetes.Clientset,
	tmpl *templates.Templates,
) *Server {
	handler := logMiddleware(log, newHandler(log, clientset, tmpl))

	return &Server{
		s: &http.Server{
			Addr:    ":8080",
			Handler: handler,
		},
	}
}

func (s *Server) Run(ctx context.Context) error {
	sderrch := make(chan error, 1)

	go func() {
		defer close(sderrch)

		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.s.Shutdown(ctx); err != nil {
			sderrch <- fmt.Errorf("shutdown: %w", err)
		}
	}()

	lerr := s.s.ListenAndServe()
	if lerr != http.ErrServerClosed {
		lerr = fmt.Errorf("listen and serve: %w", lerr)
	}

	sderr := <-sderrch

	return errors.Join(sderr, lerr)
}
