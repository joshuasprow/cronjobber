package routes

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/joshuasprow/cronjobber/pkg/k8s"
	"github.com/joshuasprow/cronjobber/pkg/models"
	"github.com/joshuasprow/cronjobber/pkg/templates"
	"github.com/joshuasprow/cronjobber/pkg/templates/components"
	"k8s.io/client-go/kubernetes"
)

type Job struct {
	log       *slog.Logger
	clientset *kubernetes.Clientset
	tmpl      *templates.Templates
}

func NewJob(
	log *slog.Logger,
	clientset *kubernetes.Clientset,
	tmpl *templates.Templates,
) Job {
	return Job{
		log:       log,
		clientset: clientset,
		tmpl:      tmpl,
	}
}

func (j Job) GET(w http.ResponseWriter, r *http.Request) {
	job, err := components.ParseJob(r)
	if err != nil {
		j.log.Error("parse job", "err", err)
		job.Error = "bad request"
	}

	if err := j.tmpl.Render(w, "components/job", j); err != nil {
		j.log.Error("execute template", "err", err)
		return
	}
}

func (j Job) POST(w http.ResponseWriter, r *http.Request) {
	job := components.Job{}
	job.Namespace = r.FormValue("namespace")
	job.OwnerName = r.FormValue("cronJobName")
	job.Name = models.FormatJobName(job.OwnerName, time.Now().UnixMilli())
	job.State = "adding"

	j.log.Info("triggering job", "cronjob", job.OwnerName, "jobname", job.Name)

	_, err := k8s.RunCronJobNow(
		r.Context(),
		j.clientset,
		job.Namespace,
		job.OwnerName,
		job.Name,
	)
	if err != nil {
		j.log.Error("run job", "jobname", job.Name, "err", err)
		job.Error = "internal server error"
	} else {
		j.log.Info("job triggered", "jobname", job.Name)
	}

	if err := j.tmpl.Render(w, "components/job", job); err != nil {
		j.log.Error("execute template", "err", err)
		return
	}
}

func (j Job) DELETE(w http.ResponseWriter, r *http.Request) {
	render := func(job components.Job) {
		if err := j.tmpl.Render(w, "components/job", job); err != nil {
			j.log.Error("execute template", "err", err)
		}
	}

	job := components.Job{}

	request, err := models.RequestFromReader(r.Body)
	if err != nil {
		j.log.Error("parse request", "err", err)
		job.Error = "bad request"
		render(job)
		return
	}

	job, err = components.ParseJob(request)
	if err != nil {
		j.log.Error("parse job", "err", err)
		job.Error = "bad request"
		render(job)
		return
	}

	j.log.Info("deleting job", "jobname", job.Name)

	if err := k8s.DeleteJob(
		r.Context(),
		j.clientset,
		job.Namespace,
		job.Name,
	); err != nil {
		j.log.Error("delete job", "jobname", job.Name, "err", err)
		job.Error = "internal server error"
	} else {
		j.log.Info("job deleted", "jobname", job.Name)
	}

	job.State = "deleting"

	if err := j.tmpl.Render(w, "components/job", job); err != nil {
		j.log.Error("execute template", "err", err)
		return
	}
}
