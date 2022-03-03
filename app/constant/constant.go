package constant

import (
	"time"
	"xenotification/app/env"
)

var (
	CORSDomain = []string{env.Config.App.SystemPath, "*"}
)

const (
	RetryAttemptCount    = 3
	RetryAttemptDuration = 1 * time.Minute
)
