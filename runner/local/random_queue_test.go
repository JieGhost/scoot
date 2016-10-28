package local

import (
	"context"
	"sync"
	"testing"

	"github.com/scootdev/scoot/runner"
	"time"
)

func TestRandomQueueing(t *testing.T) {

	runnerAvailableCh := make(chan struct{})
	testEnv, err := setup(runnerAvailableCh, false)
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer teardown(testEnv)

	ctx := context.TODO()

	qr := NewQueuingRunner(ctx, testEnv.commandRunner, 10, runnerAvailableCh).(*QueueingRunner)

	testEnv.waitGroup.Add(1)       // set the pause condition on the execer's wait group
	commandWG := &sync.WaitGroup{} // create the command specific wait group
	commandWG.Add(1)

	// start it running
	status, _ := assertQueuedRun("command wait group:", []string{"pause", "complete 0"}, commandWG, []runner.ProcessState{runner.PREPARING, runner.PENDING, runner.RUNNING}, "", qr, t)

	testEnv.waitGroup.Done() // send signal to execer's wait group

	// validate that the done signal did not terminate the run
	time.Sleep(10 * time.Millisecond)
	assertQueueStatus("execer done signal:", status.RunId, "", []runner.ProcessState{runner.RUNNING}, qr, t)

	// send the command's done signal and validate that it terminates the run
	commandWG.Done()
	time.Sleep(10 * time.Millisecond)
	assertQueueStatus("execer done signal:", status.RunId, "", []runner.ProcessState{runner.COMPLETE}, qr, t)
}
