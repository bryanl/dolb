package firewall

import (
	"bufio"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
)

var (
	iptableRuleRe = regexp.MustCompile(`^(\d+).*?ACCEPT.*?dpt:(\d+)`)
)

// PortExistsError is an error when an iptables definition exists for a port.
type PortExistsError struct {
	Port int
}

func (e *PortExistsError) Error() string {
	return fmt.Sprintf("rule for port %d exists in iptables", e.Port)
}

// Firewall is an interface for controlling a firewall.
type Firewall interface {
	Open(port int) error
	Close(port int) error
	State() (State, error)
}

// IptablesFirewall manages iptables firewalls.
type IptablesFirewall struct {
	log *logrus.Entry
	ic  IptablesCommand
}

var _ Firewall = &IptablesFirewall{}

// NewIptablesFirewall creates an instance of IptablesFirewall.
func NewIptablesFirewall(ic IptablesCommand, log *logrus.Entry) *IptablesFirewall {
	return &IptablesFirewall{
		log: log,
		ic:  ic,
	}
}

// State is current state of the firewall.
func (f *IptablesFirewall) State() (State, error) {
	out, err := f.ic.ListRules()
	if err != nil {
		return nil, err
	}

	return NewIptablesState(string(out))
}

// Open opens a port on the firewall.
func (f *IptablesFirewall) Open(port int) error {
	_, err := f.findRuleByPort(port)
	if err == nil {
		return &PortExistsError{Port: port}
	}

	return f.ic.PrependRule(port)
}

// Close closes a port on the firewall.
func (f *IptablesFirewall) Close(port int) error {
	_, err := f.findRuleByPort(port)
	if err != nil {
		return err
	}

	return f.ic.RemoveRule(port)
}

func (f *IptablesFirewall) findRuleByPort(port int) (*Rule, error) {
	state, err := f.State()
	if err != nil {
		return nil, err
	}

	rules, err := state.Rules()
	if err != nil {
		return nil, err
	}

	for _, rule := range rules {
		if rule.Destination == port {
			return &rule, nil
		}
	}

	return nil, fmt.Errorf("unable to find port %d in iptables", port)
}

// State is interface for returning firewall rules.
type State interface {
	Rules() ([]Rule, error)
}

// IptablesState reads iptables to determine the current state.
type IptablesState struct {
	in      string
	matcher *regexp.Regexp
}

var _ State = &IptablesState{}

// NewIptablesState creates an instance of IptablesState.
func NewIptablesState(in string) (*IptablesState, error) {
	return &IptablesState{
		in:      in,
		matcher: iptableRuleRe,
	}, nil
}

// Rules returns a list of rules as defined in iptables.
func (is *IptablesState) Rules() ([]Rule, error) {
	r := strings.NewReader(is.in)
	scanner := bufio.NewScanner(r)

	rules := []Rule{}

	for scanner.Scan() {
		m := is.matcher.FindAllStringSubmatch(scanner.Text(), -1)

		if m != nil {
			ruleNo, err := strconv.Atoi(m[0][1])
			if err != nil {
				return nil, err
			}

			port, err := strconv.Atoi(m[0][2])
			if err != nil {
				return nil, err
			}

			rule := Rule{
				RuleNumber:  ruleNo,
				Destination: port,
			}
			rules = append(rules, rule)
		}
	}

	return rules, nil
}

// Rule is a firewall rule. Currently the implementation only knows about TCP ports.
type Rule struct {
	// RuleNumber is the iptables rule number in a chain.
	RuleNumber int

	// Destination is the TCP destiation port.
	Destination int
}
