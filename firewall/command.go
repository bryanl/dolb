package firewall

import (
	"os/exec"
	"strconv"
)

var (
	// iptablesCmd is the iptables bin location.
	iptablesCmd = "/sbin/iptables"
)

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
	return e.cmd.CombinedOutput()
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

type IptablesCommand interface {
	PrependRule(port int) error
	RemoveRule(rule int) error
	ListRules() ([]byte, error)
}

type iptablesCommand struct {
	ExecFactory ExecFactory
}

var _ IptablesCommand = &iptablesCommand{}

func NewIptablesCommand() IptablesCommand {
	return &iptablesCommand{
		ExecFactory: &LiveExecFactory{},
	}
}

func (ic *iptablesCommand) PrependRule(port int) error {
	portStr := strconv.Itoa(port)
	opts := []string{
		"-I", "Firewall-INPUT", "1",
		"-m", "conntrack",
		"--ctstate", "NEW",
		"-p", "tcp",
		"--dport", portStr,
		"-j", "ACCEPT",
	}

	cmd := ic.newCmd(opts...)
	_, err := cmd.Exec()
	return err
}

func (ic *iptablesCommand) RemoveRule(ruleNumber int) error {
	rule := strconv.Itoa(ruleNumber)
	opts := []string{"-D", "Firewall-INPUT", rule}

	cmd := ic.newCmd(opts...)
	_, err := cmd.Exec()
	return err
}

func (ic *iptablesCommand) ListRules() ([]byte, error) {
	opts := []string{"-nL", "Firewall-INPUT", "--line-numbers"}
	cmd := ic.newCmd(opts...)
	return cmd.Exec()
}

func (ic *iptablesCommand) newCmd(opts ...string) Execer {
	return ic.ExecFactory.NewCmd(iptablesCmd, opts...)
}
