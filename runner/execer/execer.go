package execer

import (
	"io"

	"github.com/scootdev/scoot/runner"
)

// Execer lets you run one Unix command. It differs from Runner in that it does not
// know about Snapshots or Scoot. It's just a way to run a Unix process (or fake it).
// It's at the level of os/exec, not exec-as-a-service.

type CommandI interface {
	GetArgv() []string
	GetStdout() io.Writer
	GetStderr() io.Writer
}

type Command struct {
	Argv []string
	Dir  string

	Stdout io.Writer
	Stderr io.Writer
	// TODO(dbentley): environment variables?
}

func (c Command) GetArgv() []string {
	return c.Argv
}

func (c Command) GetStdout() io.Writer {
	return c.Stdout
}

func (c Command) GetStderr() io.Writer {
	return c.Stderr
}

type ProcessState int

const (
	UNKNOWN ProcessState = iota
	RUNNING
	COMPLETE
	FAILED
)

func (s ProcessState) IsDone() bool {
	return s == COMPLETE || s == FAILED
}

type Execer interface {
	// Starts process to exec command in a new goroutine.
	Exec(command CommandI) (Process, error)
	MakeExecerCommand(runnerCommand runner.CommandI, dir string, stdout, stderr io.Writer) CommandI
}

type Process interface {
	// TODO(dbentley): perhaps have a poll method?

	// Blocks until the process terminates.
	Wait() ProcessStatus

	// Terminates process and does best effort to get ExitCode.
	Abort() ProcessStatus
}

type ProcessStatus struct {
	State    ProcessState
	ExitCode int
	Error    string
}
