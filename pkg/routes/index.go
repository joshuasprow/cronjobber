package routes

import (
	"log/slog"
	"net/http"

	"github.com/joshuasprow/cronjobber/pkg/k8s"
	"github.com/joshuasprow/cronjobber/pkg/templates"
	"github.com/joshuasprow/cronjobber/pkg/templates/pages"
	"k8s.io/client-go/kubernetes"
)

type Index struct {
	log       *slog.Logger
	clientset *kubernetes.Clientset
	tmpl      *templates.Templates
}

func NewIndex(
	log *slog.Logger,
	clientset *kubernetes.Clientset,
	tmpl *templates.Templates,
) Index {
	return Index{
		log:       log,
		clientset: clientset,
		tmpl:      tmpl,
	}
}

func (i Index) GET(w http.ResponseWriter, r *http.Request) {
	index := pages.Index{
		Namespace: "default",
	}

	name := "pages/index"

	if r.Header.Get("Hx-Request") == "true" {
		// prevent browser-level caching of partial page
		w.Header().Set("Vary", "Hx-Request")

		name = "components/index"

		var err error

		index.CronJobs, err = k8s.GetCronJobs(
			r.Context(),
			i.clientset,
			index.Namespace,
		)
		if err != nil {
			i.log.Error("get cron jobs", "err", err)
			index.Error = "internal server error"
		}

		index.Loaded = true
	}

	if err := i.tmpl.Render(w, name, index); err != nil {
		i.log.Error("render template", "err", err)
		return
	}
}
