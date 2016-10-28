package execers

import (
	"github.com/scootdev/scoot/runner"
	"github.com/scootdev/scoot/runner/execer"
	"io"
)

// InterceptExecer is a Composite Execer. For each command c, if
// Condition(c), InterceptExecer delegates to Interceptor, otherwise Default
type InterceptExecer struct {
	Condition   func(execer.CommandI) bool
	Interceptor execer.Execer
	Default     execer.Execer
}

func (e *InterceptExecer) Exec(command execer.CommandI) (execer.Process, error) {
	if e.Condition(command) {
		return e.Interceptor.Exec(command)
	} else {
		return e.Default.Exec(command)
	}
}

func (e *InterceptExecer) MakeExecerCommand(runnerCommand runner.CommandI, dir string, stdout, stderr io.Writer) execer.CommandI {
	//execerCommand := ExecerCommandWithWaitGroup{}
	//execerCommand.Argv = runnerCommand.GetArgv()
	//execerCommand.Wait = runnerCommand.(runner.RunnerCommandWithWaitGroup).Wait
	//execerCommand.Dir = dir
	//execerCommand.Stdout = stdout
	//execerCommand.Stderr = stderr
	//return execerCommand
	return nil
}

// Returns whether cmd's first arg starts with UseSimExecerArg
func StartsWithSimExecer(cmd execer.CommandI) bool {

	return len(cmd.GetArgv()) > 0 && cmd.GetArgv()[0] == UseSimExecerArg
}

// A placeholder string that indicates a command should be run on SimExecer
const UseSimExecerArg = "#! sim execer"

// Create an InterceptExecer that will send cmd's to simExecer or delegate
func MakeSimExecerInterceptor(simExecer, delegate execer.Execer) execer.Execer {
	return &InterceptExecer{
		Condition:   StartsWithSimExecer,
		Interceptor: simExecer,
		Default:     delegate,
	}
}
