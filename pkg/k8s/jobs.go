package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"sort"

	"github.com/joshuasprow/cronjobber/pkg/models"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func jobToModel(job batchv1.Job) (models.Job, error) {
	def, err := json.Marshal(job)
	if err != nil {
		return models.Job{}, fmt.Errorf("marshal job: %w", err)
	}

	startTime := ""
	if job.Status.StartTime != nil {
		startTime = job.Status.StartTime.Format("2006/01/02 15:04:05")
	}
	completionTime := ""
	if job.Status.CompletionTime != nil {
		completionTime = job.Status.CompletionTime.Format("2006/01/02 15:04:05")
	}

	return models.Job{
		Namespace:      job.Namespace,
		OwnerUID:       string(job.OwnerReferences[0].UID),
		OwnerName:      job.OwnerReferences[0].Name,
		Name:           job.Name,
		StartTime:      startTime,
		CompletionTime: completionTime,
		Succeeded:      job.Status.Succeeded > 0,
		Def:            string(def),
	}, nil
}

func GetJob(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	namespace string,
	jobName string,
) (
	models.Job,
	bool,
	error,
) {
	job, err := clientset.
		BatchV1().
		Jobs(namespace).
		Get(ctx, jobName, metav1.GetOptions{})
	if serr, ok := err.(*errors.StatusError); ok &&
		serr.ErrStatus.Code == http.StatusNotFound {
		return models.Job{}, false, nil
	}
	if err != nil {
		return models.Job{}, false, fmt.Errorf("get job: %w", err)
	}
	if job == nil {
		return models.Job{}, false, fmt.Errorf("job is nil")
	}

	model, err := jobToModel(*job)
	if err != nil {
		return models.Job{}, true, fmt.Errorf("job to model: %w", err)
	}

	return model, true, nil
}

func GetJobs(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	namespace string,
	cronJobUid string,
) (
	[]models.Job,
	error,
) {
	list, err := clientset.
		BatchV1().
		Jobs(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}

	jobs := []models.Job{}

	for _, item := range list.Items {
		if !slices.ContainsFunc(
			item.OwnerReferences,
			func(r metav1.OwnerReference) bool {
				return string(r.UID) == cronJobUid
			},
		) {
			continue
		}

		model, err := jobToModel(item)
		if err != nil {
			return nil, fmt.Errorf("job to model: %w", err)
		}

		jobs = append(jobs, model)
	}

	sort.SliceStable(jobs, func(i, j int) bool {
		return jobs[i].Name > jobs[j].Name
	})

	return jobs, nil
}

func DeleteJob(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	namespace string,
	name string,
) error {
	propagationPolicy := metav1.DeletePropagationBackground

	if err := clientset.
		BatchV1().
		Jobs(namespace).
		Delete(
			ctx,
			name,
			metav1.DeleteOptions{PropagationPolicy: &propagationPolicy},
		); err != nil {
		return fmt.Errorf("delete job: %w", err)
	}

	return nil
}
