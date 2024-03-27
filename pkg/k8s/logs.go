package k8s

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/joshuasprow/cronjobber/pkg"
	"github.com/joshuasprow/cronjobber/pkg/models"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func StreamPodLogs(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	container models.Container,
	logsCh chan<- pkg.Result[string],
) {
	defer close(logsCh)

	req := clientset.
		CoreV1().
		Pods(container.Namespace).
		GetLogs(
			container.Pod,
			&v1.PodLogOptions{
				Container: container.Name,
				Follow:    true,
				TailLines: pkg.Pointer(int64(10)),
			},
		)

	stream, err := req.Stream(ctx)
	if err != nil {
		logsCh <- pkg.Result[string]{Err: fmt.Errorf("get stream: %w", err)}
		return
	}
	defer func() {
		if err := stream.Close(); err != nil {
			logsCh <- pkg.Result[string]{Err: fmt.Errorf("close stream: %w", err)}
		}
	}()

	scanner := bufio.NewScanner(stream)

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			logsCh <- pkg.Result[string]{Err: fmt.Errorf("scan error: %w", err)}
			return
		}

		logsCh <- pkg.Result[string]{V: strings.TrimSpace(scanner.Text())}
	}
}
