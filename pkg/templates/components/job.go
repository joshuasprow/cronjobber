package components

import (
	"errors"

	"github.com/joshuasprow/cronjobber/pkg/models"
)

type Job struct {
	models.Job

	Error string
}

func ParseJob(request models.Request) (Job, error) {
	j := Job{
		Job: models.Job{
			Namespace:      request.FormValue("namespace"),
			OwnerName:      request.FormValue("ownerName"),
			Name:           request.FormValue("jobName"),
			StartTime:      request.FormValue("startTime"),
			CompletionTime: request.FormValue("completionTime"),
			Succeeded:      request.FormValue("succeeded") == "true",
		},
	}

	errs := []error{}
	if j.Namespace == "" {
		errs = append(errs, errors.New("namespace is empty"))
	}
	if j.OwnerName == "" {
		errs = append(errs, errors.New("ownerName is empty"))
	}
	if j.Name == "" {
		errs = append(errs, errors.New("name is empty"))
	}

	if err := errors.Join(errs...); err != nil {
		return Job{}, err
	}

	return j, nil
}
