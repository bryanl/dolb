package firewall

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIptablesState(t *testing.T) {
	Convey("reading state", t, func() {
		is, err := NewIptablesState(state)
		So(err, ShouldBeNil)

		rules, err := is.Rules()
		So(err, ShouldBeNil)

		So(rules, ShouldHaveLength, 1)
		So(rules[0].Destination, ShouldEqual, 8889)
	})
}

var state = `Chain Firewall-INPUT (2 references)
num  target     prot opt source               destination         
1    ACCEPT     tcp  --  0.0.0.0/0            0.0.0.0/0            ctstate NEW tcp dpt:8889
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
12   ACCEPT     tcp  --  0.0.0.0/0            0.0.0.0/0            ctstate NEW multiport dports 22,80,443
13   LOG        all  --  0.0.0.0/0            0.0.0.0/0            LOG flags 0 level 4
14   REJECT     all  --  0.0.0.0/0            0.0.0.0/0            reject-with icmp-port-unreachable
`
