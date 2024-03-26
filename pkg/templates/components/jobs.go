package components

import (
	"errors"
	"net/http"

	"github.com/joshuasprow/cronjobber/pkg/models"
)

type Jobs struct {
	Namespace  string
	CronJobUid string
	Jobs       []models.Job
	Error      string
}

func ParseJobs(request *http.Request) (Jobs, error) {
	j := Jobs{
		Namespace:  request.FormValue("namespace"),
		CronJobUid: request.FormValue("cronJobUid"),
	}

	if j.Namespace == "" && j.CronJobUid == "" {
		query := request.URL.Query()

		j.Namespace = query.Get("namespace")
		j.CronJobUid = query.Get("cronJobUid")
	}

	errs := []error{}

	if j.Namespace == "" {
		errs = append(errs, errors.New("namespace is empty"))
	}
	if j.CronJobUid == "" {
		return Jobs{}, errors.New("cronJobUid is empty")
	}

	if err := errors.Join(errs...); err != nil {
		return Jobs{}, err
	}

	return j, nil
}
