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
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func mockIPProviders(t *testing.T, opendnsResponse string, ipinfoResponse string) (*url.URL, func()) {
	// mock opendns provider
	servers, serverAddresses := startMockDNSServerFleet(t, map[string][]string{"myip.opendns.com.": []string{opendnsResponse}})

	tmpOpenDNSNameservers := openDNSNameservers
	openDNSNameservers = serverAddresses

	// mock ipinfo provider
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprintf(w, fmt.Sprintf(`{"ip": "%s", "hostname": "", "city": "", "region": "", "country": "", "loc": "", "org": ""}`, ipinfoResponse))
	}))
	mockURL, _ := url.Parse(ts.URL)

	return mockURL, func() {
		stopMockDNSServerFleet(servers)
		openDNSNameservers = tmpOpenDNSNameservers
		ts.Close()
	}
}

func TestIPProviderCombined_Get(t *testing.T) {
	mockURL, cleanup := mockIPProviders(t, "1.1.1.1", "1.1.1.1")
	defer cleanup()

	p := newIPProviderCombined()
	p.ipinfo.url = mockURL

	// test combined provider
	ip, err := p.Get()
	if err != nil {
		t.Fatalf("combined ip provider returned unexpected error: %+v", err)
	}

	if !ip.Equal(net.ParseIP("1.1.1.1")) {
		t.Fatalf("combined ip provider returned unexpected response")
	}
}

func TestIPProviderCombined_Get_errIPInfo(t *testing.T) {
	p := newIPProviderCombined()

	// mock opendns provider
	servers, serverAddresses := startMockDNSServerFleet(t, map[string][]string{"myip.opendns.com.": []string{"1.1.1.1"}})

	tmpOpenDNSNameservers := openDNSNameservers
	openDNSNameservers = serverAddresses

	defer func() {
		openDNSNameservers = tmpOpenDNSNameservers
		stopMockDNSServerFleet(servers)
	}()

	// mock ipinfo provider
	u, _ := url.Parse("https://127.0.0.1:10000")
	p.ipinfo.url = u

	// test combined provider
	ip, err := p.Get()
	if err != nil {
		t.Fatalf("combined ip provider returned unexpected error: %+v", err)
	}

	if !ip.Equal(net.ParseIP("1.1.1.1")) {
		t.Fatalf("combined ip provider returned unexpected response")
	}
}

func TestIPProviderCombined_Get_errOpenDNS(t *testing.T) {
	p := newIPProviderCombined()

	// mock ipinfo provider
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprintf(w, `{"ip": "1.1.1.1", "hostname": "", "city": "", "region": "", "country": "", "loc": "", "org": ""}`)
	}))

	// mock opendns provider
	tmpOpenDNSNameservers := openDNSNameservers
	openDNSNameservers = []string{"127.0.0.1:10000"}
	defer func() {
		ts.Close()
		openDNSNameservers = tmpOpenDNSNameservers
	}()

	mockURL, _ := url.Parse(ts.URL)
	p.ipinfo.url = mockURL

	// test combined provider
	ip, err := p.Get()
	if err != nil {
		t.Fatalf("combined ip provider returned unexpected error: %+v", err)
	}

	if !ip.Equal(net.ParseIP("1.1.1.1")) {
		t.Fatalf("combined ip provider returned unexpected response")
	}
}

func TestIPProviderCombined_Get_errBoth(t *testing.T) {
	p := newIPProviderCombined()

	// mock opendns provider
	tmpOpenDNSNameservers := openDNSNameservers
	openDNSNameservers = []string{"127.0.0.1:10000"}
	defer func() {
		openDNSNameservers = tmpOpenDNSNameservers
	}()

	mockURL, _ := url.Parse("https://127.0.0.1:10000")
	p.ipinfo.url = mockURL

	// test combined provider
	_, err := p.Get()
	if err != errCombinedBothFailed {
		t.Fatalf("combined ip provider did not return expected error")
	}
}

func TestIPProviderCombined_Get_errDifferent(t *testing.T) {
	mockURL, cleanup := mockIPProviders(t, "1.1.1.1", "1.2.3.4")
	defer cleanup()

	p := newIPProviderCombined()
	p.ipinfo.url = mockURL

	// test combined provider
	_, err := p.Get()
	if err != errCombinedDifferentResults {
		t.Fatalf("combined ip provider did not return expected error")
	}
}
