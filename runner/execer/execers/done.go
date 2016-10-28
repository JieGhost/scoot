package execers

import (
	"github.com/scootdev/scoot/runner"
	"github.com/scootdev/scoot/runner/execer"
	"io"
)

// Creates a new doneExecer.
func NewDoneExecer() execer.Execer {
	return &doneExecer{}
}

// doneExecer finishes something as soon as its run
type doneExecer struct{}

func (e *doneExecer) Exec(command execer.CommandI) (execer.Process, error) {
	return e, nil
}

func (e *doneExecer) MakeExecerCommand(runnerCommand runner.CommandI, dir string, stdout, stderr io.Writer) execer.CommandI {
	return nil
}

var completeStatus = execer.ProcessStatus{
	State:    execer.COMPLETE,
	ExitCode: 0,
}

func (e *doneExecer) Wait() execer.ProcessStatus {
	return completeStatus
}

func (e *doneExecer) Abort() execer.ProcessStatus {
	return completeStatus
}
