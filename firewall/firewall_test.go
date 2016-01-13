package firewall

import (
	"testing"

	"github.com/Sirupsen/logrus"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIptablesFirewallOpen(t *testing.T) {

	Convey("Given an instance of IptablesFirewall", t, func() {
		var (
			log = logrus.WithFields(logrus.Fields{})
			err error
		)

		ic := &MockIptablesCommand{}
		fw := NewIptablesFirewall(ic, log)

		Convey("When rule does not already exist", func() {
			ic.On("ListRules").Return([]byte(output), nil)
			ic.On("PrependRule", 80).Return(nil)

			err = fw.Open(80)

			Convey("Then it opens a port using iptables", func() {
				So(err, ShouldBeNil)
				ic.AssertExpectations(t)
			})

		})

		Convey("When rule already exists", func() {
			ic.On("ListRules").Return([]byte(output80exists), nil)

			err = fw.Open(80)

			Convey("It returns a PortExistsError", func() {
				if perr, ok := err.(*PortExistsError); ok {
					So(perr.Port, ShouldEqual, 80)
				} else {
					t.Errorf("unexpected error: %v")
				}
			})
		})
	})
}

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
