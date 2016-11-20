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

package dns

import (
	"net"
	"testing"

	"github.com/alkar/odyn/dnstest"
)

func Test_resolveARecord_noServer(t *testing.T) {
	dc := NewClient()
	_, err := dc.ResolveARecord("example.com.", []string{"127.0.0.1:65111"})
	if err == nil {
		t.Fatalf("resolveARecord should have returned an error")
	}
}

func Test_resolveARecord_empty(t *testing.T) {
	servers, serverAddresses, err := dnstest.StartMockDNSServerFleet(map[string][]string{"example.com.": []string{}})
	defer dnstest.StopMockDNSServerFleet(servers)
	if err != nil {
		t.Fatalf("unable to run test server: %v", err)
	}

	dc := NewClient()
	_, err = dc.ResolveARecord("example.com.", serverAddresses)
	if err != ErrEmptyAnswer {
		t.Fatalf("resolveARecord should have returned an empty answer error")
	}
}

func Test_resolveARecord_multipleDifferent(t *testing.T) {
	servers, serverAddresses, err := dnstest.StartMockDNSServerFleet(map[string][]string{"example.com.": []string{"1.1.1.1", "1.2.3.4"}})
	defer dnstest.StopMockDNSServerFleet(servers)
	if err != nil {
		t.Fatalf("unable to run test server: %v", err)
	}

	dc := NewClient()
	resp, err := dc.ResolveARecord("example.com.", serverAddresses)
	if err != nil {
		t.Fatalf("resolveARecord returned unexpected error: %+v", err)
	}

	if len(resp) != 2 {
		t.Fatalf("resolveARecord should have returned two values")
	}

	if !resp[0].Equal(net.ParseIP("1.1.1.1")) || !resp[1].Equal(net.ParseIP("1.2.3.4")) {
		t.Fatalf("resolveARecord returned unexpected response")
	}
}

func Test_resolveARecord_multipleSame(t *testing.T) {
	servers, serverAddresses, err := dnstest.StartMockDNSServerFleet(map[string][]string{"example.com.": []string{"1.1.1.1", "1.1.1.1"}})
	defer dnstest.StopMockDNSServerFleet(servers)
	if err != nil {
		t.Fatalf("unable to run test server: %v", err)
	}

	dc := NewClient()
	resp, err := dc.ResolveARecord("example.com.", serverAddresses)
	if err != nil {
		t.Fatalf("resolveARecord returned unexpected error: %+v", err)
	}

	if len(resp) != 1 {
		t.Fatalf("resolveARecord should return a single value if the response contains the same value multiple times")
	}

	if !resp[0].Equal(net.ParseIP("1.1.1.1")) {
		t.Fatalf("resolveARecord returned unexpected response")
	}
}

func Test_resolveARecord_broken(t *testing.T) {
	servers, serverAddresses, err := dnstest.StartMockSemiBrokenDNSServerFleet(map[string][]string{"example.com.": []string{"1.1.1.1", "1.1.1.1"}})
	defer dnstest.StopMockDNSServerFleet(servers)
	if err != nil {
		t.Fatalf("unable to run test server: %v", err)
	}

	dc := NewClient()
	resp, err := dc.ResolveARecord("example.com.", serverAddresses)
	if err != nil {
		t.Fatalf("resolveARecord returned unexpected error: %+v", err)
	}

	if len(resp) != 1 {
		t.Fatalf("resolveARecord should return a single value if the response contains the same value multiple times")
	}

	if !resp[0].Equal(net.ParseIP("1.1.1.1")) {
		t.Fatalf("resolveARecord returned unexpected response")
	}
}
