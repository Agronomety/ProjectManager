package models

import (
	"time"
)

type Project struct {
	ID          int64
	Name        string
	Path        string
	Description string
	ReadmePath  string
	LastOpened  time.Time
	Tags        []string
	Icon        string
}
