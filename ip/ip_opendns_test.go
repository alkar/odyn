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
	"net"
	"testing"
)

func TestIPProviderOpenDNS_Get(t *testing.T) {
	servers, serverAddresses := startMockDNSServerFleet(t, map[string][]string{"myip.opendns.com.": []string{"1.1.1.1"}})
	defer stopMockDNSServerFleet(servers)

	// mock opendns servers
	tmpOpenDNSNameservers := openDNSNameservers
	openDNSNameservers = serverAddresses
	defer func() {
		openDNSNameservers = tmpOpenDNSNameservers
	}()

	p := newIPProviderOpenDNS()

	ip, err := p.Get()
	if err != nil {
		t.Fatalf("resolveARecord returned unexpected error: %+v", err)
	}

	if !ip.Equal(net.ParseIP("1.1.1.1")) {
		t.Fatalf("resolveARecord returned unexpected response")
	}
}

func TestIPProviderOpenDNS_Get_broken(t *testing.T) {
	servers, serverAddresses := startMockSemiBrokenDNSServerFleet(t, map[string][]string{"myip.opendns.com.": []string{"1.1.1.1"}})
	defer stopMockDNSServerFleet(servers)

	// mock opendns servers
	tmpOpenDNSNameservers := openDNSNameservers
	openDNSNameservers = serverAddresses
	defer func() {
		openDNSNameservers = tmpOpenDNSNameservers
	}()

	p := newIPProviderOpenDNS()

	ip, err := p.Get()
	if err != nil {
		t.Fatalf("resolveARecord returned unexpected error: %+v", err)
	}

	if !ip.Equal(net.ParseIP("1.1.1.1")) {
		t.Fatalf("resolveARecord returned unexpected response")
	}
}

func TestIPProviderOpenDNS_Get_error(t *testing.T) {
	// mock opendns servers
	tmpOpenDNSNameservers := openDNSNameservers
	openDNSNameservers = []string{"127.0.0.1:65111"}
	defer func() {
		openDNSNameservers = tmpOpenDNSNameservers
	}()

	p := newIPProviderOpenDNS()

	_, err := p.Get()
	if err == nil {
		t.Fatalf("resolveARecord should have returned an error")
	}
}

func TestIPProviderOpenDNS_Get_multipleDifferent(t *testing.T) {
	servers, serverAddresses := startMockDNSServerFleet(t, map[string][]string{"myip.opendns.com.": []string{"1.1.1.1", "1.2.3.4"}})
	defer stopMockDNSServerFleet(servers)

	// mock opendns servers
	tmpOpenDNSNameservers := openDNSNameservers
	openDNSNameservers = serverAddresses
	defer func() {
		openDNSNameservers = tmpOpenDNSNameservers
	}()

	p := newIPProviderOpenDNS()

	ip, err := p.Get()
	if err != nil {
		t.Fatalf("resolveARecord returned unexpected error: %+v", err)
	}

	if !ip.Equal(net.ParseIP("1.1.1.1")) {
		t.Fatalf("resolveARecord returned unexpected response")
	}
}
