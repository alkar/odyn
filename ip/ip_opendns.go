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

package ip

import (
	"errors"
	"log"
	"net"
)

const (
	openDNSTargetHostname = "myip.opendns.com."
)

var (
	errOpenDNSEmptyAnswer = errors.New("OpenDNS answer was empty")

	openDNSNameservers = []string{
		"208.67.222.222:53", // resolver1.opendns.com
		"208.67.220.220:53", // resolver2.opendns.com
		"208.67.222.220:53", // resolver3.opendns.com
		"208.67.220.222:53", // resolver4.opendns.com
	}
)

type ipProviderOpenDNS struct {
	dnsClient *dnsClient
}

func newIPProviderOpenDNS() *ipProviderOpenDNS {
	return &ipProviderOpenDNS{
		dnsClient: newDNSClient(),
	}
}

func (p ipProviderOpenDNS) Get() (net.IP, error) {
	ips, err := p.dnsClient.resolveARecord(openDNSTargetHostname, openDNSNameservers)
	if err != nil {
		return nil, err
	}

	if len(ips) != 1 {
		log.Printf("DNS answer contained multiple entries, will only use the first: %+v", ips)
	}

	return ips[0], err
}
