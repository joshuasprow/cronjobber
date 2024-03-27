package routes

import (
	"log/slog"
	"net/http"

	"github.com/joshuasprow/cronjobber/pkg/k8s"
	"github.com/joshuasprow/cronjobber/pkg/templates"
	"github.com/joshuasprow/cronjobber/pkg/templates/pages"
	"k8s.io/client-go/kubernetes"
)

type CronJobDef struct {
	log       *slog.Logger
	clientset *kubernetes.Clientset
	tmpl      *templates.Templates
}

func NewCronJobDef(
	log *slog.Logger,
	clientset *kubernetes.Clientset,
	tmpl *templates.Templates,
) CronJobDef {
	return CronJobDef{
		log:       log,
		clientset: clientset,
		tmpl:      tmpl,
	}
}

func (j CronJobDef) GET(w http.ResponseWriter, r *http.Request) {
	render := func(c pages.CronJob) {
		if err := j.tmpl.Render(w, "components/cronjobdef", c); err != nil {
			j.log.Error("execute template", "err", err)
		}
	}

	component, err := pages.ParseCronJob(r)
	if err != nil {
		j.log.Error("parse cron job", "err", err)
		component.Error = "bad request"
		render(component)
		return
	}

	cronJob, err := k8s.GetCronJob(
		r.Context(),
		j.clientset,
		component.Namespace,
		component.Name,
	)
	if err != nil {
		j.log.Error("get cron job", "err", err)
		component.Error = "internal server error"
		render(component)
		return
	}

	component.CronJob = cronJob

	render(component)
}
