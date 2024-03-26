package pages

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/joshuasprow/cronjobber/pkg/models"
)

type CronJob struct {
	models.CronJob

	Jobs   []models.Job // todo: remove placeholder for nested jobs template
	Loaded bool
	Error  string
}

func parseFormCronJob(request models.Request) models.CronJob {
	return models.CronJob{
		Namespace:          request.FormValue("namespace"),
		Name:               request.FormValue("cronJobName"),
		Schedule:           request.FormValue("schedule"),
		LastScheduleTime:   request.FormValue("lastScheduleTime"),
		LastSuccessfulTime: request.FormValue("lastSuccessfulTime"),
	}
}

func parseQueryCronJob(request *http.Request) models.CronJob {
	return models.CronJob{
		Namespace:          request.URL.Query().Get("namespace"),
		Name:               request.URL.Query().Get("name"),
		Schedule:           request.URL.Query().Get("schedule"),
		LastScheduleTime:   request.URL.Query().Get("lastScheduleTime"),
		LastSuccessfulTime: request.URL.Query().Get("lastSuccessfulTime"),
	}
}

func validateCronJob(c models.CronJob, source string) (models.CronJob, error) {
	errs := []error{}

	if c.Namespace == "" {
		errs = append(errs, fmt.Errorf("%s: namespace is empty", source))
	}
	if c.Name == "" {
		errs = append(errs, fmt.Errorf("%s: name is empty", source))
	}

	if err := errors.Join(errs...); err != nil {
		return c, err
	}

	return c, nil
}

func ParseCronJob(request *http.Request) (CronJob, error) {
	fcj, ferr := validateCronJob(parseFormCronJob(request), "form")
	if ferr == nil {
		return CronJob{CronJob: fcj}, nil
	}

	qcj, qerr := validateCronJob(parseQueryCronJob(request), "query")
	if qerr == nil {
		return CronJob{CronJob: fcj}, nil
	}

	return CronJob{CronJob: qcj}, errors.Join(ferr, qerr)
}
