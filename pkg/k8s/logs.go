package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/joshuasprow/cronjobber/pkg"
	"github.com/joshuasprow/cronjobber/pkg/models"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func parseLogLineTimestamp(timestamp string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, timestamp)
	if err == nil {
		return t, nil
	}

	t, err = time.Parse(time.RFC3339, timestamp)
	if err == nil {
		return t, nil
	}

	return time.Time{}, err
}

func readLogLine(line string) (models.Log, error) {
	parts := strings.SplitN(line, " ", 2)

	if len(parts) != 2 {
		return models.Log{}, fmt.Errorf("invalid format: %q", line)
	}

	t, err := parseLogLineTimestamp(string(parts[0]))
	if err != nil {
		return models.Log{}, fmt.Errorf("parse timestamp: %w", err)
	}

	return models.Log{
		Timestamp: t,
		Message:   string(parts[1]),
	}, nil
}

func newGetLogsRequest(
	clientset *kubernetes.Clientset,
	container models.Container,
	follow bool,
) *rest.Request {
	tailLines := int64(100)
	if follow {
		tailLines = 1
	}
	return clientset.
		CoreV1().
		Pods(container.Namespace).
		GetLogs(
			container.Pod,
			&v1.PodLogOptions{
				Container:  container.Name,
				Follow:     follow,
				TailLines:  &tailLines,
				Timestamps: true,
			},
		)
}

func GetLogs(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	container models.Container,
) (
	[]models.Log,
	error,
) {
	body, err := newGetLogsRequest(clientset, container, false).DoRaw(ctx)
	if err != nil {
		return nil, fmt.Errorf("get logs: %w", err)
	}

	logs := []models.Log{}

	for _, data := range bytes.Split(body, []byte("\n")) {
		line := strings.TrimSpace(string(data))

		if line == "" {
			continue
		}

		log, err := readLogLine(line)
		if err != nil {
			return nil, fmt.Errorf("read log line: %w", err)
		}

		logs = append(logs, log)
	}

	return logs, nil
}

type GetLogResult pkg.Result[models.Log]

func StreamLogs(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	container models.Container,
	logsCh chan<- GetLogResult,
) {
	defer close(logsCh)

	check := func(msg string, err error) bool {
		if err != nil {
			logsCh <- GetLogResult{Err: fmt.Errorf("%s: %w", msg, err)}
			return true
		}
		return false
	}

	stream, err := newGetLogsRequest(clientset, container, true).Stream(ctx)
	if check("get stream:", err) {
		return
	}
	defer func() { check("close stream", stream.Close()) }()

	buf := make([]byte, 16*1024) // 16KB: the maximum size of a log line in k8s

	for {
		n, err := stream.Read(buf)
		if err == io.EOF {
			return
		}
		if check("read stream:", err) {
			return
		}
		if n == 0 {
			return
		}

		line := strings.TrimSpace(string(buf[:n]))

		if line == "" {
			continue
		}

		log, err := readLogLine(line)
		if check("read log line:", err) {
			return
		}

		logsCh <- GetLogResult{V: log}
	}
}
