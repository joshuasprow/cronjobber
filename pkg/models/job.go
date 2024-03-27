package models

import (
	"fmt"
)

type Job struct {
	Namespace      string
	OwnerUID       string
	OwnerName      string
	Name           string
	StartTime      string
	CompletionTime string
	Succeeded      bool
	Container      Container
	State          string
	Def            string
}

func FormatJobName(cronJobName string, timestamp int64) string {
	return fmt.Sprintf("%s-manual-%d", cronJobName, timestamp)
}

type JobDef struct {
	Metadata struct {
		Name            string `json:"name"`
		Namespace       string `json:"namespace"`
		OwnerReferences []struct {
			UID  string `json:"uid"`
			Name string `json:"name"`
			Kind string `json:"kind"`
		} `json:"ownerReferences"`
	} `json:"metadata"`
	Status struct {
		StartTime      string `json:"startTime"`
		CompletionTime string `json:"completionTime"`
		Succeeded      int    `json:"succeeded"`
	} `json:"status"`
}
