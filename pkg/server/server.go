package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/joshuasprow/cronjobber/pkg/routes"
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
	mux := http.NewServeMux()

	index := routes.NewIndex(log, clientset, tmpl)
	mux.HandleFunc("GET /", index.GET)

	cronjob := routes.NewCronJob(log, clientset, tmpl)
	mux.HandleFunc("GET /cronjob", cronjob.GET)

	cronjobdef := routes.NewCronJobDef(log, clientset, tmpl)
	mux.HandleFunc("GET /cronjobdef", cronjobdef.GET)

	jobs := routes.NewJobs(log, clientset, tmpl)
	mux.HandleFunc("GET /jobs", jobs.GET)

	job := routes.NewJob(log, clientset, tmpl)
	mux.HandleFunc("GET /job", job.GET)
	mux.HandleFunc("POST /job", job.POST)
	mux.HandleFunc("DELETE /job", job.DELETE)

	jobdef := routes.NewJobDef(log, clientset, tmpl)
	mux.HandleFunc("GET /jobdef", jobdef.GET)

	return &Server{
		s: &http.Server{
			Addr:    ":8080",
			Handler: logMiddleware(log, mux),
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
