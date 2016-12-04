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
	"net"
	"testing"
)

func TestDNSProvider_Get(t *testing.T) {
	servers, serverAddresses, err := startMockDNSServerFleet(map[string][]string{"myip.opendns.com.": []string{"1.1.1.1"}})
	defer stopMockDNSServerFleet(servers)
	if err != nil {
		t.Fatalf("unable to run test server: %v", err)
	}

	p, _ := NewDNSProvider("myip.opendns.com.", serverAddresses)

	ip, err := p.Get()
	if err != nil {
		t.Fatalf("DNSProvider.Get returned unexpected error: %+v", err)
	}

	if !ip.Equal(net.ParseIP("1.1.1.1")) {
		t.Fatalf("DNSProvider.Get returned unexpected response")
	}
}

func TestDNSProvider_Get_broken(t *testing.T) {
	servers, serverAddresses, err := startMockSemiBrokenDNSServerFleet(map[string][]string{"myip.opendns.com.": []string{"1.1.1.1"}})
	defer stopMockDNSServerFleet(servers)
	if err != nil {
		t.Fatalf("unable to run test server: %v", err)
	}

	p, _ := NewDNSProvider("myip.opendns.com.", serverAddresses)

	ip, err := p.Get()
	if err != nil {
		t.Fatalf("DNSProvider.Get returned unexpected error: %+v", err)
	}

	if !ip.Equal(net.ParseIP("1.1.1.1")) {
		t.Fatalf("DNSProvider.Get returned unexpected response")
	}
}

func TestDNSProvider_Get_noResults(t *testing.T) {
	p, _ := NewDNSProvider("myip.opendns.com.", nil)

	_, err := p.Get()
	if err != ErrDNSProviderNoResults {
		t.Fatalf("DNSProvider.Get did not return expected error")
	}
}

func TestDNSProvider_Get_error(t *testing.T) {
	p, _ := NewDNSProvider("myip.opendns.com.", []string{"127.0.0.1:64321"})

	_, err := p.Get()
	if err == nil {
		t.Fatalf("DNSProvider.Get did was expected to return an error")
	}
}

func TestDNSProvider_Get_multipleDifferent(t *testing.T) {
	servers, serverAddresses, err := startMockDNSServerFleet(map[string][]string{"myip.opendns.com.": []string{"1.1.1.1", "1.2.3.4"}})
	defer stopMockDNSServerFleet(servers)
	if err != nil {
		t.Fatalf("unable to run test server: %v", err)
	}

	p, _ := NewDNSProvider("myip.opendns.com.", serverAddresses)

	ip, err := p.Get()
	if err != ErrDNSProviderMultipleResults {
		t.Fatalf("DNSProvider.Get returned unexpected error: %+v", err)
	}

	if !ip.Equal(net.ParseIP("1.1.1.1")) {
		t.Fatalf("DNSProvider.Get returned unexpected response")
	}
}
