package routes

import (
	"log/slog"
	"net/http"

	"github.com/joshuasprow/cronjobber/pkg/k8s"
	"github.com/joshuasprow/cronjobber/pkg/templates"
	"github.com/joshuasprow/cronjobber/pkg/templates/components"
	"k8s.io/client-go/kubernetes"
)

type Jobs struct {
	log       *slog.Logger
	clientset *kubernetes.Clientset
	tmpl      *templates.Templates
}

func NewJobs(
	log *slog.Logger,
	clientset *kubernetes.Clientset,
	tmpl *templates.Templates,
) Jobs {
	return Jobs{
		log:       log,
		clientset: clientset,
		tmpl:      tmpl,
	}
}

func (j Jobs) GET(w http.ResponseWriter, r *http.Request) {
	jobs, err := components.ParseJobs(r)
	if err != nil {
		j.log.Error("parse jobs", "err", err)
		jobs.Error = "bad request"

		if err := j.tmpl.Render(w, "components/jobs", jobs); err != nil {
			j.log.Error("render template", "err", err)
			return
		}

		return
	}

	jobs.Jobs, err = k8s.GetJobs(
		r.Context(),
		j.clientset,
		jobs.Namespace,
		jobs.CronJobUid,
	)
	if err != nil {
		j.log.Error("get jobs", "err", err)
		jobs.Error = "internal server error"
	}

	if err := j.tmpl.Render(w, "components/jobs", jobs); err != nil {
		j.log.Error("render template", "err", err)
		return
	}
}
