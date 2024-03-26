package routes

import (
	"log/slog"
	"net/http"

	"github.com/joshuasprow/cronjobber/pkg/k8s"
	"github.com/joshuasprow/cronjobber/pkg/templates"
	"github.com/joshuasprow/cronjobber/pkg/templates/pages"
	"k8s.io/client-go/kubernetes"
)

type CronJob struct {
	log       *slog.Logger
	clientset *kubernetes.Clientset
	tmpl      *templates.Templates
}

func NewCronJob(
	log *slog.Logger,
	clientset *kubernetes.Clientset,
	tmpl *templates.Templates,
) CronJob {
	return CronJob{
		log:       log,
		clientset: clientset,
		tmpl:      tmpl,
	}
}

func (c CronJob) GET(w http.ResponseWriter, r *http.Request) {
	cronJob, err := pages.ParseCronJob(r)
	if err != nil {
		c.log.Error("parse cron job", "err", err)
		cronJob.Error = "bad request"
	}

	name := "pages/cronjob"

	if r.Header.Get("Hx-Request") == "true" {
		// prevent browser-level caching of partial page
		w.Header().Set("Vary", "Hx-Request")

		name = "components/cronjob"

		cronJob.Loaded = true

		cronJob.CronJob, err = k8s.GetCronJob(
			r.Context(),
			c.clientset,
			cronJob.Namespace,
			cronJob.Name,
		)
		if err != nil {
			c.log.Error("get cron jobs", "err", err)
			cronJob.Error = "internal server error"

			if err := c.tmpl.Render(w, name, cronJob); err != nil {
				c.log.Error("execute template", "err", err)
				return
			}

			return
		}
	}

	if err := c.tmpl.Render(w, name, cronJob); err != nil {
		c.log.Error("execute template", "err", err)
		return
	}
}
