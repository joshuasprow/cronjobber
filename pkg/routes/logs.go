package routes

import (
	"log/slog"
	"net/http"

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
	logs, err := pages.ParseLogs(r)
	if err != nil {
		l.log.Error("parse cron job", "err", err)
		logs.Error = "bad request"
	}

	name := "pages/logs"

	// if r.Header.Get("Hx-Request") == "true" {
	// 	// prevent browser-level caching of partial page
	// 	w.Header().Set("Vary", "Hx-Request")

	// 	name = "components/logs"

	// 	logs.Loaded = true

	// 	logs.Logs, err = k8s.GetLogs(
	// 		r.Context(),
	// 		l.clientset,
	// 		logs.Namespace,
	// 		logs.Name,
	// 	)
	// 	if err != nil {
	// 		l.log.Error("get cron jobs", "err", err)
	// 		logs.Error = "internal server error"

	// 		if err := l.tmpl.Render(w, name, logs); err != nil {
	// 			l.log.Error("execute template", "err", err)
	// 			return
	// 		}

	// 		return
	// 	}
	// }

	if err := l.tmpl.Render(w, name, logs); err != nil {
		l.log.Error("execute template", "err", err)
		return
	}
}
