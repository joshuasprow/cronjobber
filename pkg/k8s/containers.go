package k8s

import (
	"context"
	"fmt"

	"github.com/joshuasprow/cronjobber/pkg/models"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
)

func SetJobsContainers(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	namespace string,
	jobs []models.Job,
) (
	[]models.Job,
	error,
) {
	names := []string{}

	for _, job := range jobs {
		names = append(names, job.Name)
	}

	req, err := labels.NewRequirement("job-name", selection.In, names)
	if err != nil {
		return nil, fmt.Errorf("new label requirement: %w", err)
	}

	labelSelector := labels.NewSelector().Add(*req).String()

	pods, err := GetPods(ctx, clientset, namespace, labelSelector)
	if err != nil {
		return nil, fmt.Errorf("get pods: %w", err)
	}

	for _, pod := range pods {
		if len(pod.OwnerReferences) != 1 {
			return nil, fmt.Errorf("expected pod %s to have 1 owner, got %d", pod.Name, len(pod.OwnerReferences))
		}

		if len(pod.Spec.Containers) == 0 {
			continue
		}

		if len(pod.Spec.Containers) > 1 {
			return nil, fmt.Errorf("expected pod %s to have 0 or 1 container, got %d", pod.Name, len(pod.Spec.Containers))
		}

		o := pod.OwnerReferences[0]
		c := pod.Spec.Containers[0]

		for i, job := range jobs {
			if job.Name == o.Name {
				jobs[i].Container = models.Container{
					Name:      c.Name,
					Namespace: pod.Namespace,
					Pod:       pod.Name,
				}
				break
			}
		}
	}

	return jobs, nil
}

func GetPods(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	namespace string,
	labelSelector string,
) (
	[]v1.Pod,
	error,
) {
	list, err := clientset.
		CoreV1().
		Pods(namespace).
		List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}
