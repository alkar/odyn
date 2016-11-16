package main

import (
	"errors"
	"net"

	"github.com/miekg/dns"
)

const (
	defaultDNSTTLSeconds = 60
)

var (
	errDNSEmptyAnswer = errors.New("DNS nameserver returned an empty answer")
)

type dnsClient struct {
	*dns.Client
}

func newDNSClient() *dnsClient {
	return &dnsClient{&dns.Client{}}
}

func (c *dnsClient) resolveARecord(name string, nameservers []string) ([]net.IP, error) {
	m := dns.Msg{}
	m.SetQuestion(name, dns.TypeA)

	var retError error
	var retIP []net.IP

	for _, nameserver := range nameservers {
		r, _, err := c.Exchange(&m, nameserver)
		if err != nil {
			retError = err
			continue
		}

		if len(r.Answer) == 0 {
			retError = errDNSEmptyAnswer
			continue
		}

		for _, ans := range r.Answer {
			ip := ans.(*dns.A).A

			exists := false
			for _, i := range retIP {
				if ip.Equal(i) {
					exists = true
					break
				}
			}

			if !exists {
				retIP = append(retIP, ip)
			}
		}

		retError = nil
		break
	}

	return retIP, retError
}
