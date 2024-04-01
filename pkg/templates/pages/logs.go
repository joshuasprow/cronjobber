package pages

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/joshuasprow/cronjobber/pkg/models"
)

type Logs struct {
	models.Container

	Logs   []models.Log
	Loaded bool
	Error  string
}

func parseFormLogs(request models.Request) models.Container {
	return models.Container{
		Namespace: request.FormValue("namespace"),
		Pod:       request.FormValue("pod"),
		Name:      request.FormValue("name"),
	}
}

func parseQueryLogs(request *http.Request) models.Container {
	return models.Container{
		Namespace: request.URL.Query().Get("namespace"),
		Pod:       request.URL.Query().Get("pod"),
		Name:      request.URL.Query().Get("name"),
	}
}

func validateLogs(c models.Container, source string) (models.Container, error) {
	errs := []error{}

	if c.Namespace == "" {
		errs = append(errs, fmt.Errorf("%s: namespace is empty", source))
	}
	if c.Pod == "" {
		errs = append(errs, fmt.Errorf("%s: pod is empty", source))
	}
	if c.Name == "" {
		errs = append(errs, fmt.Errorf("%s: name is empty", source))
	}

	if err := errors.Join(errs...); err != nil {
		return c, err
	}

	return c, nil
}

func ParseLogs(request *http.Request) (Logs, error) {
	fl, ferr := validateLogs(parseFormLogs(request), "form")
	if ferr == nil {
		return Logs{Container: fl}, nil
	}

	ql, qerr := validateLogs(parseQueryLogs(request), "query")
	if qerr == nil {
		return Logs{Container: fl}, nil
	}

	return Logs{Container: ql}, errors.Join(ferr, qerr)
}
