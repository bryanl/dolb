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
		fw  = NewIptablesFirewall(ef, fs, log)
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

				b := []byte("output")

				e.On("Exec").Return(b, nil).Once()

				ef.On("NewCmd", "/usr/sbin/iptables", []string{"-I", "Firewall-INPUT", "1", "-m", "conntrack", "--ctstate", "NEW", "-p", "tcp", "--dport", "80", "-j", "ACCEPT"}).Return(e, nil).Once()

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

				b := []byte("output")

				e.On("Exec").Return(b, nil).Once()

				ef.On("NewCmd", "/usr/sbin/iptables", []string{"-D", "Firewall-INPUT", "2"}).Return(e, nil).Once()
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
			})

			It("returns an error because the port is not opened", func() {
				Ω(err).To(HaveOccurred())
			})
		})

	})

})
