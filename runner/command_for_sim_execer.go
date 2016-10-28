package runner

import (
	"sync"
	"time"
)

// for testing only!
type RunnerCommandWithWaitGroup struct {
	Command
	Wait *sync.WaitGroup
}

func (c RunnerCommandWithWaitGroup) GetTimeout() time.Duration {
	return c.Timeout
}
func (c RunnerCommandWithWaitGroup) GetSnapshotId() string {
	return c.SnapshotId
}
func (c RunnerCommandWithWaitGroup) GetArgv() []string {
	return c.Argv
}
