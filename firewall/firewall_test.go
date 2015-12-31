package firewall_test

import (
	"github.com/Sirupsen/logrus"
	. "github.com/bryanl/dolb/firewall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IptablesFirewall", func() {

	var (
		err error
		log = logrus.WithFields(logrus.Fields{})
		e   = &MockExecer{}
		ef  = &MockExecFactory{}
		fs  = &MockFirewallState{}
		fw  = NewIptablesFirewall(ef, log)
	)

	Describe("Open", func() {

		JustBeforeEach(func() {
			err = fw.Open(80)
		})

		Context("rule does not already exist", func() {

			BeforeEach(func() {
				rules := []Rule{
					{Destination: 22, RuleNumber: 1},
				}
				fs.On("Rules").Return(rules, nil).Once()

				b := []byte(output)
				e.On("Exec").Return(b, nil).Once()

				ef.On("NewCmd", IptablesCmd, []string{"-I", "Firewall-INPUT", "1", "-m", "conntrack", "--ctstate", "NEW", "-p", "tcp", "--dport", "80", "-j", "ACCEPT"}).Return(e, nil).Once()

				e.On("Exec").Return(b, nil).Once()
				ef.On("NewCmd", IptablesCmd, []string{"-nL", "Firewall-INPUT", "--line-numbers"}).Return(e, nil).Once()

			})

			It("opens a port using iptables", func() {
				Ω(err).ToNot(HaveOccurred())

				e.AssertExpectations(GinkgoT())
				ef.AssertExpectations(GinkgoT())
			})

		})

		Context("rule already exists", func() {

			BeforeEach(func() {
				rules := []Rule{
					{Destination: 22, RuleNumber: 1},
					{Destination: 80, RuleNumber: 2},
				}
				fs.On("Rules").Return(rules, nil).Once()

				b := []byte(output80exists)
				e.On("Exec").Return(b, nil).Once()
				ef.On("NewCmd", IptablesCmd, []string{"-nL", "Firewall-INPUT", "--line-numbers"}).Return(e, nil).Once()
			})

			It("opens a port using iptables", func() {
				Ω(err).To(HaveOccurred())

				e.AssertExpectations(GinkgoT())
				ef.AssertExpectations(GinkgoT())
			})

		})
	})

	Describe("Close", func() {

		JustBeforeEach(func() {
			err = fw.Close(80)
		})

		Context("port that is already opened", func() {

			BeforeEach(func() {
				rules := []Rule{
					{Destination: 22, RuleNumber: 1},
					{Destination: 80, RuleNumber: 2},
				}
				fs.On("Rules").Return(rules, nil).Once()

				b := []byte(output80exists)
				e.On("Exec").Return(b, nil).Once()
				ef.On("NewCmd", IptablesCmd, []string{"-nL", "Firewall-INPUT", "--line-numbers"}).Return(e, nil).Once()

				b = []byte("")
				e.On("Exec").Return(b, nil).Once()
				ef.On("NewCmd", IptablesCmd, []string{"-D", "Firewall-INPUT", "1"}).Return(e, nil).Once()

			})

			It("removes an iptables entry for a port", func() {
				Ω(err).ToNot(HaveOccurred())

				e.AssertExpectations(GinkgoT())
				ef.AssertExpectations(GinkgoT())
			})

		})

		Context("port is not opened", func() {
			BeforeEach(func() {
				rules := []Rule{
					{Destination: 22, RuleNumber: 1},
				}
				fs.On("Rules").Return(rules, nil).Once()

				b := []byte(output)
				e.On("Exec").Return(b, nil).Once()
				ef.On("NewCmd", IptablesCmd, []string{"-nL", "Firewall-INPUT", "--line-numbers"}).Return(e, nil).Once()

			})

			It("returns an error because the port is not opened", func() {
				Ω(err).To(HaveOccurred())
			})
		})

	})

})

var output80exists = `Chain Firewall-INPUT (2 references)
num  target     prot opt source               destination
1    ACCEPT     tcp  --  0.0.0.0/0            0.0.0.0/0            ctstate NEW tcp dpt:80
2    ACCEPT     all  --  0.0.0.0/0            0.0.0.0/0
3    ACCEPT     icmp --  0.0.0.0/0            0.0.0.0/0            icmptype 0
4    ACCEPT     icmp --  0.0.0.0/0            0.0.0.0/0            icmptype 3
5    ACCEPT     icmp --  0.0.0.0/0            0.0.0.0/0            icmptype 11
6    ACCEPT     icmp --  0.0.0.0/0            0.0.0.0/0            icmptype 8
7    ACCEPT     all  --  0.0.0.0/0            0.0.0.0/0            ctstate RELATED,ESTABLISHED
8    ACCEPT     all  --  10.137.227.148       0.0.0.0/0
9    ACCEPT     all  --  10.137.19.148        0.0.0.0/0
10   ACCEPT     all  --  10.137.99.148        0.0.0.0/0
11   ACCEPT     all  --  0.0.0.0/0            0.0.0.0/0
13   LOG        all  --  0.0.0.0/0            0.0.0.0/0            LOG flags 0 level 4
14   REJECT     all  --  0.0.0.0/0            0.0.0.0/0            reject-with icmp-port-unreachable`

var output = `Chain Firewall-INPUT (2 references)
num  target     prot opt source               destination
2    ACCEPT     all  --  0.0.0.0/0            0.0.0.0/0
3    ACCEPT     icmp --  0.0.0.0/0            0.0.0.0/0            icmptype 0
4    ACCEPT     icmp --  0.0.0.0/0            0.0.0.0/0            icmptype 3
5    ACCEPT     icmp --  0.0.0.0/0            0.0.0.0/0            icmptype 11
6    ACCEPT     icmp --  0.0.0.0/0            0.0.0.0/0            icmptype 8
7    ACCEPT     all  --  0.0.0.0/0            0.0.0.0/0            ctstate RELATED,ESTABLISHED
8    ACCEPT     all  --  10.137.227.148       0.0.0.0/0
9    ACCEPT     all  --  10.137.19.148        0.0.0.0/0
10   ACCEPT     all  --  10.137.99.148        0.0.0.0/0
11   ACCEPT     all  --  0.0.0.0/0            0.0.0.0/0
13   LOG        all  --  0.0.0.0/0            0.0.0.0/0            LOG flags 0 level 4
14   REJECT     all  --  0.0.0.0/0            0.0.0.0/0            reject-with icmp-port-unreachable`
