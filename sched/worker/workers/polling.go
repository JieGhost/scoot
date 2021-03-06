package workers

import (
	"time"

	"github.com/scootdev/scoot/runner"
	"github.com/scootdev/scoot/sched"
	"github.com/scootdev/scoot/sched/worker"
)

// NewPollingWorker creates a PollingWorker
func NewPollingWorker(
	runner runner.Runner,
	period time.Duration) worker.Worker {
	return NewPollingWorkerWithTimeout(runner, period, false, 0*time.Minute)
}

// Creates a New Polling Worker which enforces a Timeout on Tasks
func NewPollingWorkerWithTimeout(
	runner runner.Runner,
	period time.Duration,
	enforceTimout bool,
	timeout time.Duration) worker.Worker {
	return &PollingWorker{
		runner:         runner,
		pollingPeriod:  period,
		enforceTimeout: enforceTimout,
		timeout:        timeout,
	}
}

// PollingWorker acts as a Worker by polling the underlying runner every period
type PollingWorker struct {
	runner         runner.Runner
	pollingPeriod  time.Duration
	enforceTimeout bool
	timeout        time.Duration
}

func (r *PollingWorker) RunAndWait(task sched.TaskDefinition) (runner.ProcessStatus, error) {
	// schedule the task
	status, err := r.runner.Run(&task.Command)
	if err != nil {
		return status, err
	}

	timeSpent := 0 * time.Second
	id := status.RunId

	// Periodically check in with the worker to get a task update
	for !r.enforceTimeout || (r.enforceTimeout && timeSpent < r.timeout) {
		time.Sleep(r.pollingPeriod)
		timeSpent += r.pollingPeriod

		status, err = r.runner.Status(id)
		if err != nil || status.State.IsDone() {
			return status, err
		}
	}

	// The Task took to long to run. Tell the runner to abort before returning
	status.State = runner.TIMEDOUT
	_, err = r.runner.Abort(status.RunId)
	return status, err
}
