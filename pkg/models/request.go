package models

import (
	"fmt"
	"io"
	"net/url"
)

type Request interface {
	FormValue(key string) string
}

type readerRequest struct {
	values url.Values
}

func (r readerRequest) FormValue(key string) string {
	return r.values.Get(key)
}

// This function exists because Go doesn't parse form data from a DELETE request by default.
func RequestFromReader(reader io.Reader) (Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read request body: %w", err)

	}

	values, err := url.ParseQuery(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse query: %w", err)
	}

	return readerRequest{values: values}, nil
}
