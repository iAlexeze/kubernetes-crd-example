package utils

import "time"

type Status string

const (
	// Status
	StatusReady   Status = "ready"
	StatusRunning Status = "running"
	StatusHealthy Status = "healthy"
	StatusOnline  Status = "online"

	StatusNotReady   Status = "not ready"
	StatusNotRunning Status = "not running"
	StatusNotHealthy Status = "not healthy"
	StatusOffline    Status = "offline"

	// HTTP
	ContentType     = "Content-Type"
	JSONContentType = "application/json"
)

type H map[string]interface{}

func Sleep(n int) {
	time.Sleep(time.Duration(n) * time.Second)
}
