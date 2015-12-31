package firewall

import "os/exec"

// Execer is implemented by any values that has a Exec() method. The Exec method
// is used to run a command return the output or an error.
type Execer interface {
	Exec() ([]byte, error)
}

// LiveExecer execs a command directly.
type LiveExecer struct {
	cmd *exec.Cmd
}

var _ Execer = &LiveExecer{}

// Exec execs a command returns the output or an error.
func (e *LiveExecer) Exec() ([]byte, error) {
	return e.cmd.Output()
}

// ExecFactory is an interface for a factory than can created commands.
type ExecFactory interface {
	NewCmd(name string, args ...string) Execer
}

// LiveExecFactory returns instances of LiveExecer.
type LiveExecFactory struct {
}

var _ ExecFactory = &LiveExecFactory{}

// NewCmd creates a new command to be ran.
func (f *LiveExecFactory) NewCmd(name string, args ...string) Execer {
	return &LiveExecer{
		cmd: exec.Command(name, args...),
	}
}
