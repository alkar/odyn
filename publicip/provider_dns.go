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

package publicip

import (
	"errors"
	"net"

	"github.com/alkar/odyn/dns"
)

var (
	// ErrDNSProviderNoResults is returned when no results could be obtained
	// from the DNS nameservers.
	ErrDNSProviderNoResults = errors.New("dns provider returned no results")

	// ErrDNSProviderMultipleResults is returned when multiple different IP
	// addresses where obtained from the set of nameservers. In this case, the
	// provider will, however, return the first IP address instead of nil.
	ErrDNSProviderMultipleResults = errors.New("dns provider returned multiple different results")
)

// DNSProvider sends queries to a DNS nameserver to discover the public IP
// address.
type DNSProvider struct {
	dns         *dns.Client
	record      string
	nameservers []string
}

// NewDNSProvider returns an instantiated DNSProvider.
func NewDNSProvider(record string, nameservers []string) (*DNSProvider, error) {
	return &DNSProvider{
		dns:         dns.NewClient(),
		record:      record,
		nameservers: nameservers,
	}, nil
}

// Get performs a DNS query and returns the IP address in the answer.
func (p DNSProvider) Get() (net.IP, error) {
	ips, err := p.dns.ResolveARecord(p.record, p.nameservers)
	if err != nil {
		return nil, err
	}

	if len(ips) == 0 {
		return nil, ErrDNSProviderNoResults
	}

	if len(ips) > 1 {
		return ips[0], ErrDNSProviderMultipleResults
	}

	return ips[0], nil
}
