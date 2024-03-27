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

var logCh chan pkg.Result[string]

func (l Logs) GET(w http.ResponseWriter, r *http.Request) {
	logs, err := pages.ParseLogs(r)
	if err != nil {
		l.log.Error("parse cron job", "err", err)
		logs.Error = "bad request"
	}

	if r.Header.Get("Accept") == "text/event-stream" {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		flusher, ok := w.(http.Flusher)
		if !ok {
			l.log.Error("streaming logs", "err", "streaming unsupported")
			http.Error(w, "streaming unsupported", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		if logCh == nil {
			logCh = make(chan pkg.Result[string])
			go k8s.StreamPodLogs(ctx, l.clientset, logs.Container, logCh)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case log, ok := <-logCh:
				if log.Err != nil {
					l.log.Error("streaming logs", "err", log.Err)
					return
				}
				if !ok {
					return
				}

				data := fmt.Sprintf("event: message\ndata: <tr><td><pre>%s</pre></td></tr>\n\n", log.V)

				if _, err := w.Write([]byte(data)); err != nil {
					l.log.Error("write log message", "err", err)
					return
				}

				flusher.Flush()
			}
		}
	}

	if err := l.tmpl.Render(w, "pages/logs", logs); err != nil {
		l.log.Error("execute template", "err", err)
		return
	}
}
