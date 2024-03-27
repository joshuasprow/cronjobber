package routes

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/joshuasprow/cronjobber/pkg"
	"github.com/joshuasprow/cronjobber/pkg/k8s"
	"github.com/joshuasprow/cronjobber/pkg/templates"
	"github.com/joshuasprow/cronjobber/pkg/templates/pages"
	"k8s.io/client-go/kubernetes"
)

type Logs struct {
	log       *slog.Logger
	clientset *kubernetes.Clientset
	tmpl      *templates.Templates
}

func NewLogs(
	log *slog.Logger,
	clientset *kubernetes.Clientset,
	tmpl *templates.Templates,
) Logs {
	return Logs{
		log:       log,
		clientset: clientset,
		tmpl:      tmpl,
	}
}

func (l Logs) GET(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Header.Get("Accept") == "text/event-stream":
		fmt.Println("GET /logs (stream)")
		if err := l.handleGetStream(w, r); err != nil {
			l.log.Error("handle get stream", "err", err)
		}
	case r.Header.Get("Hx-Request") == "true":
		fmt.Println("GET /logs (component)")
		// prevent browser-level caching of partial page
		// w.Header().Set("Vary", "Hx-Request")

		if err := l.handleGetComponent(w, r); err != nil {
			l.log.Error("handle get component", "err", err)
		}
	default:
		fmt.Println("GET /logs (page)")
		if err := l.handleGetPage(w, r); err != nil {
			l.log.Error("handle get page", "err", err)
		}
	}
}

func (l Logs) handleGetComponent(w http.ResponseWriter, r *http.Request) error {
	logs, err := pages.ParseLogs(r)
	if err != nil {
		return fmt.Errorf("parse logs: %w", err)
	}

	messages, err := k8s.GetLogs(r.Context(), l.clientset, logs.Container)
	if err != nil {
		return fmt.Errorf("get logs: %w", err)
	}

	logs.Messages = messages
	logs.Loaded = true

	if err := l.tmpl.Render(w, "components/logs", logs); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	return nil
}

func (l Logs) handleGetPage(w http.ResponseWriter, r *http.Request) error {
	logs, err := pages.ParseLogs(r)
	if err != nil {
		return fmt.Errorf("parse logs: %w", err)
	}

	if err := l.tmpl.Render(w, "pages/logs", logs); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	return nil
}

func (l Logs) handleGetStream(w http.ResponseWriter, r *http.Request) error {
	logs, err := pages.ParseLogs(r)
	if err != nil {
		return fmt.Errorf("parse logs: %w", err)
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("http flusher not supported")
	}
	defer flusher.Flush()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	logCh := make(chan pkg.Result[string])

	go k8s.StreamLogs(ctx, l.clientset, logs.Container, logCh)

	for {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("context done: %w", err)
			}

			return nil
		case log := <-logCh:
			if log.Err != nil {
				return fmt.Errorf("stream logs: %w", log.Err)
			}
			if log.V == "" {
				break
			}
			if _, err := w.Write([]byte("event: message\n")); err != nil {
				return fmt.Errorf("write event: %w", err)
			}
			if _, err := w.Write([]byte("data: <tr><td><pre>" + log.V + "</pre></td></tr>\n\n")); err != nil {
				return fmt.Errorf("write data: %w", err)
			}
			flusher.Flush()
		}
	}
}
