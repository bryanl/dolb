package firewall

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_iptablesCommand(t *testing.T) {

	successCmd := func() *MockExecer {
		e := &MockExecer{}
		e.On("Exec").Return([]byte("output"), nil)
		return e
	}

	Convey("Given an iptablesCommand", t, func() {
		execFactory := &MockExecFactory{}
		ic := &iptablesCommand{
			ExecFactory: execFactory,
		}

		Convey("When prepending a rule", func() {
			cmd := successCmd()
			execFactory.On(
				"NewCmd",
				"/sbin/iptables",
				[]string{"-I", "Firewall-INPUT", "1", "-m", "conntrack", "--ctstate", "NEW", "-p", "tcp", "--dport", "80", "-j", "ACCEPT"},
			).Return(cmd)

			err := ic.PrependRule(80)

			Convey("Then it adds the rule without an error", func() {
				So(err, ShouldBeNil)
				cmd.AssertExpectations(t)
			})
		})

		Convey("When removing a rule", func() {
			cmd := successCmd()
			execFactory.On(
				"NewCmd",
				"/sbin/iptables",
				[]string{"-D", "Firewall-INPUT", "1"},
			).Return(cmd)

			err := ic.RemoveRule(1)

			Convey("Then it removes the rule without an error", func() {
				So(err, ShouldBeNil)
				cmd.AssertExpectations(t)
			})
		})

		Convey("When listing rules", func() {
			cmd := successCmd()
			execFactory.On(
				"NewCmd",
				"/sbin/iptables",
				[]string{"-nL", "Firewall-INPUT", "--line-numbers"},
			).Return(cmd)

			output, err := ic.ListRules()

			Convey("It returns no error", func() {
				So(err, ShouldBeNil)
			})

			Convey("It returns the command output", func() {
				So(string(output), ShouldResemble, "output")
			})
		})
	})
}

func TestLiveExecer(t *testing.T) {
	Convey("Given a LiveExecFactory", t, func() {
		factory := &LiveExecFactory{}

		Convey("When creating a command", func() {
			cmd := factory.NewCmd("id")
			output, err := cmd.Exec()

			Convey("It returns no error", func() {
				So(err, ShouldBeNil)
			})

			Convey("It returns the command output", func() {
				So(string(output), ShouldStartWith, "uid=")
			})
		})
	})
}
