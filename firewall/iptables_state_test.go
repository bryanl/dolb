package firewall_test

import (
	. "github.com/bryanl/dolb/firewall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IptablesState", func() {

	Describe("reading state", func() {

		var (
			is  *IptablesState
			err error
		)

		BeforeEach(func() {
			is, err = NewIptablesState(state)
			Ω(err).ToNot(HaveOccurred())
		})

		It("has 1 applicable rule", func() {
			rules, err := is.Rules()
			Ω(err).ToNot(HaveOccurred())

			Ω(len(rules)).To(Equal(1))
			Ω(rules[0].Destination).To(Equal(8889))
		})
	})
})

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
