package routes

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/joshuasprow/cronjobber/pkg/k8s"
	"github.com/joshuasprow/cronjobber/pkg/templates"
	"github.com/joshuasprow/cronjobber/pkg/templates/components"
	"k8s.io/client-go/kubernetes"
)

type JobDef struct {
	log       *slog.Logger
	clientset *kubernetes.Clientset
	tmpl      *templates.Templates
}

func NewJobDef(
	log *slog.Logger,
	clientset *kubernetes.Clientset,
	tmpl *templates.Templates,
) JobDef {
	return JobDef{
		log:       log,
		clientset: clientset,
		tmpl:      tmpl,
	}
}

func (j JobDef) GET(w http.ResponseWriter, r *http.Request) {
	render := func(c components.Job) {
		if err := j.tmpl.Render(w, "components/jobdef", c); err != nil {
			j.log.Error("execute template", "err", err)
		}
	}

	component, err := components.ParseJob(r)
	if err != nil {
		j.log.Error("parse job", "err", err)
		component.Error = "bad request"
		render(component)
		return
	}

	job, found, err := k8s.GetJob(
		r.Context(),
		j.clientset,
		component.Namespace,
		component.Name,
	)
	if err != nil {
		j.log.Error("get job", "err", err)
		component.Error = "internal server error"
		render(component)
		return
	}

	if !found {
		j.log.Error("job not found", "name", component.Name)
		component.Error = fmt.Sprintf("job %s not found", component.Name)
		render(component)
		return
	}

	component.Job = job

	render(component)
}
