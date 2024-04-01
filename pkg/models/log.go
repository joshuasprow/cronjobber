package models

import "time"

type Log struct {
	Timestamp time.Time
	Message   string
}
