// Copyright 2016 Dimitrios Karagiannis
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package odyn

import (
	"errors"
	"net"

	"github.com/miekg/dns"
)

var (
	// ErrDNSEmptyAnswer is returned when the DNS client receives an empty
	// response from the nameservers.
	ErrDNSEmptyAnswer = errors.New("DNS nameserver returned an empty answer")
)

// DNSClient provides easy to use DNS resolving methods.
type DNSClient struct {
	*dns.Client
}

// NewDNSClient instantiates a new DNS client.
func NewDNSClient() *DNSClient {
	return &DNSClient{&dns.Client{}}
}

// ResolveA will ask the provided nameservers for an A record of the provided
// DNS name and return the list of IP addresses in the answer, if any.
func (c *DNSClient) ResolveA(name string, nameservers []string) ([]net.IP, error) {
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
			retError = ErrDNSEmptyAnswer
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

// DNSZone is an interface for DNS Zone providers to implement.
type DNSZone interface {
	UpdateA(recordName string, zoneName string, ip net.IP) error
	Nameservers(zoneName string) ([]string, error)
}
