package models

import batchv1 "k8s.io/api/batch/v1"

type CronJob struct {
	Namespace          string
	Uid                string
	Name               string
	Schedule           string
	LastScheduleTime   string
	LastSuccessfulTime string
	ActiveJobNames     []string
	Spec               batchv1.CronJob
	Def                string
}

type CronJobDef struct {
	Metadata struct {
		Uid       string `json:"uid"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Spec struct {
		Schedule string `json:"schedule"`
	} `json:"spec"`
	Status struct {
		LastScheduleTime   string `json:"lastScheduleTime"`
		LastSuccessfulTime string `json:"lastSuccessfulTime"`
	} `json:"status"`
}
