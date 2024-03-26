package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/joshuasprow/cronjobber/pkg/models"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

func cronJobToModel(cronJob batchv1.CronJob) (models.CronJob, error) {
	def, err := json.Marshal(cronJob)
	if err != nil {
		return models.CronJob{}, fmt.Errorf("marshal cron job: %w", err)
	}

	lastScheduleTime := ""
	if cronJob.Status.LastScheduleTime != nil {
		lastScheduleTime = cronJob.Status.LastScheduleTime.Time.Format("2006/01/02 15:04:05")
	}
	lastSuccessfulTime := ""
	if cronJob.Status.LastSuccessfulTime != nil {
		lastSuccessfulTime = cronJob.Status.LastSuccessfulTime.Time.Format("2006/01/02 15:04:05")
	}

	activeJobNames := []string{}

	for _, obj := range cronJob.Status.Active {
		if obj.Kind == "Job" {
			activeJobNames = append(activeJobNames, obj.Name)
		}
	}

	return models.CronJob{
		Namespace:          cronJob.Namespace,
		Uid:                string(cronJob.UID),
		Name:               cronJob.Name,
		Schedule:           cronJob.Spec.Schedule,
		LastScheduleTime:   lastScheduleTime,
		LastSuccessfulTime: lastSuccessfulTime,
		ActiveJobNames:     activeJobNames,
		Spec:               cronJob,
		Def:                string(def),
	}, nil
}

func GetCronJob(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	namespace string,
	name string,
) (
	models.CronJob,
	error,
) {
	cronJob, err := clientset.BatchV1().
		CronJobs(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return models.CronJob{}, fmt.Errorf("get cron job: %w", err)
	}

	if cronJob == nil {
		return models.CronJob{}, fmt.Errorf("cron job is nil")
	}

	return cronJobToModel(*cronJob)
}

func GetCronJobs(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	namespace string,
) (
	[]models.CronJob,
	error,
) {
	list, err := clientset.BatchV1().
		CronJobs(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list cron jobs: %w", err)
	}

	cronJobs := []models.CronJob{}

	for _, cronJob := range list.Items {
		model, err := cronJobToModel(cronJob)
		if err != nil {
			return nil, fmt.Errorf("cron job to model: %w", err)
		}

		cronJobs = append(cronJobs, model)
	}

	return cronJobs, nil
}

func RunCronJobNow(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	namespace string,
	cronJobName string,
	jobName string,
) (
	models.Job,
	error,
) {
	_, found, err := GetJob(ctx, clientset, namespace, jobName)
	if err != nil {
		return models.Job{}, fmt.Errorf("get job: %w", err)
	}
	if found {
		return models.Job{}, fmt.Errorf("job already exists")
	}

	cronJob, err := GetCronJob(ctx, clientset, namespace, cronJobName)
	if err != nil {
		return models.Job{}, fmt.Errorf("get cron job: %w", err)
	}

	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "batch/v1",
					Kind:       "CronJob",
					Name:       cronJobName,
					UID:        types.UID(cronJob.Uid),
				},
			},
		},
		Spec: cronJob.Spec.Spec.JobTemplate.Spec,
	}

	job, err = clientset.
		BatchV1().
		Jobs(namespace).
		Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return models.Job{}, fmt.Errorf("create job: %w", err)
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
		OwnerUID:       cronJob.Uid,
		OwnerName:      cronJob.Name,
		Name:           job.Name,
		StartTime:      startTime,
		CompletionTime: completionTime,
		Succeeded:      job.Status.Succeeded > 0,
	}, nil
}
