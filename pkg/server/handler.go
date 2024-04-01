package server

import (
	"log/slog"
	"net/http"

	"github.com/joshuasprow/cronjobber/pkg/routes"
	"github.com/joshuasprow/cronjobber/pkg/templates"
	"k8s.io/client-go/kubernetes"
)

func newHandler(
	log *slog.Logger,
	clientset *kubernetes.Clientset,
	tmpl *templates.Templates,
) http.Handler {
	h := http.NewServeMux()

	index := routes.NewIndex(log, clientset, tmpl)
	h.HandleFunc("GET /", index.GET)

	cronjob := routes.NewCronJob(log, clientset, tmpl)
	h.HandleFunc("GET /cronjob", cronjob.GET)

	cronjobdef := routes.NewCronJobDef(log, clientset, tmpl)
	h.HandleFunc("GET /cronjobdef", cronjobdef.GET)

	jobs := routes.NewJobs(log, clientset, tmpl)
	h.HandleFunc("GET /jobs", jobs.GET)

	job := routes.NewJob(log, clientset, tmpl)
	h.HandleFunc("GET /job", job.GET)
	h.HandleFunc("POST /job", job.POST)
	h.HandleFunc("DELETE /job", job.DELETE)

	jobdef := routes.NewJobDef(log, clientset, tmpl)
	h.HandleFunc("GET /jobdef", jobdef.GET)

	logs := routes.NewLogs(log, clientset, tmpl)
	h.HandleFunc("GET /logs", logs.GET)

	return h
}
