package pages

import (
	"github.com/joshuasprow/cronjobber/pkg/models"
)

type Index struct {
	Error     string
	Loaded    bool
	Namespace string
	CronJobs  []models.CronJob
}
