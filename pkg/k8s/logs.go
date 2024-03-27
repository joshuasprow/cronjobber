package k8s

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/joshuasprow/cronjobber/pkg"
	"github.com/joshuasprow/cronjobber/pkg/models"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func GetLogs(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	container models.Container,
) (
	[]string,
	error,
) {
	req, err := clientset.
		CoreV1().
		Pods(container.Namespace).
		GetLogs(
			container.Pod,
			&v1.PodLogOptions{Container: container.Name},
		).
		DoRaw(ctx)
	if err != nil {
		return nil, fmt.Errorf("get logs: %w", err)
	}

	return strings.Split(strings.TrimSpace(string(req)), "\n"), nil
}

func StreamLogs(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	container models.Container,
	logsCh chan<- pkg.Result[string],
) {
	defer close(logsCh)

	stream, err := clientset.
		CoreV1().
		Pods(container.Namespace).
		GetLogs(
			container.Pod,
			&v1.PodLogOptions{
				Container: container.Name,
				Follow:    true,
				TailLines: pkg.Pointer(int64(1)),
			},
		).
		Stream(ctx)
	if err != nil {
		logsCh <- pkg.Result[string]{Err: fmt.Errorf("get stream: %w", err)}
		return
	}
	defer func() {
		if err := stream.Close(); err != nil {
			logsCh <- pkg.Result[string]{Err: fmt.Errorf("close stream: %w", err)}
		}
	}()

	buf := make([]byte, 1024)

	for {
		n, err := stream.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			logsCh <- pkg.Result[string]{Err: fmt.Errorf("read stream: %w", err)}
			return
		}

		if n == 0 {
			break
		}

		v := strings.TrimSpace(string(buf[:n]))

		fmt.Println(v)

		logsCh <- pkg.Result[string]{V: v}
	}
}
