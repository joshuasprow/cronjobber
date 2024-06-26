package server

import (
	"bytes"
	"log/slog"
	"net/http"
	"strings"
)

type logMiddlewareWriter struct {
	http.ResponseWriter
	bodyBuf    *bytes.Buffer
	statusCode int
}

func newLogMiddlewareWriter(w http.ResponseWriter) *logMiddlewareWriter {
	return &logMiddlewareWriter{
		ResponseWriter: w,
		bodyBuf:        &bytes.Buffer{},
		statusCode:     http.StatusOK,
	}
}

func (w *logMiddlewareWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *logMiddlewareWriter) Write(b []byte) (int, error) {
	w.bodyBuf.Write(b)
	return w.ResponseWriter.Write(b)
}

func truncateString(str string, length int) string {
	body := strings.TrimSpace(str)
	if len(body) > length {
		return body[:length-3] + "..."
	}
	return body
}

func logMiddleware(log *slog.Logger, handler http.Handler) http.HandlerFunc {
	const bodyLength = 80

	return func(w http.ResponseWriter, r *http.Request) {
		args := []any{
			"method", r.Method,
			"url", truncateString(r.URL.String(), bodyLength),
		}

		if r.Method == "POST" {
			if err := r.ParseForm(); err != nil {
				log.Error("parse form", "err", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			args = append(args, "form", truncateString(r.Form.Encode(), bodyLength))
		}

		log.Info("request", args...)

		rw := newLogMiddlewareWriter(w)

		handler.ServeHTTP(rw, r)

		log.Info(
			"response",
			"status",
			rw.statusCode,
			"body",
			truncateString(rw.bodyBuf.String(), bodyLength),
		)
	}
}
