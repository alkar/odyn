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
	"sync"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func setupMockDNSRecord(mux *dns.ServeMux, name string, records []string) {
	mux.HandleFunc(name, func(w dns.ResponseWriter, req *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(req)
		m.Authoritative = true

		for _, r := range records {
			m.Answer = append(m.Answer, &dns.A{
				Hdr: dns.RR_Header{
					Name:   req.Question[0].Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				A: net.ParseIP(r).To4(),
			})
		}

		w.WriteMsg(m)
	})
}

func startMockDNSServer(laddr string, records map[string][]string) (*dns.Server, string, error) {
	pc, err := net.ListenPacket("udp", laddr)
	if err != nil {
		return nil, "", err
	}

	mux := dns.NewServeMux()
	for n, r := range records {
		setupMockDNSRecord(mux, n, r)
	}

	server := &dns.Server{
		PacketConn:   pc,
		ReadTimeout:  time.Hour,
		WriteTimeout: time.Hour,
		Handler:      mux,
	}

	waitLock := sync.Mutex{}
	waitLock.Lock()
	server.NotifyStartedFunc = waitLock.Unlock

	go func() {
		server.ActivateAndServe()
		pc.Close()
	}()

	waitLock.Lock()
	return server, pc.LocalAddr().String(), nil
}

func startMockDNSServerFleet(t *testing.T, records map[string][]string) ([]*dns.Server, []string) {
	servers := []*dns.Server{}
	serverAddresses := []string{}

	s, addr, err := startMockDNSServer("127.0.0.1:0", records)
	if err != nil {
		t.Fatalf("unable to run test server: %v", err)
	}
	servers = append(servers, s)
	serverAddresses = append(serverAddresses, addr)

	s, addr, err = startMockDNSServer("127.0.0.1:0", records)
	if err != nil {
		t.Fatalf("unable to run test server: %v", err)
	}
	servers = append(servers, s)
	serverAddresses = append(serverAddresses, addr)

	s, addr, err = startMockDNSServer("127.0.0.1:0", records)
	if err != nil {
		t.Fatalf("unable to run test server: %v", err)
	}
	servers = append(servers, s)
	serverAddresses = append(serverAddresses, addr)

	s, addr, err = startMockDNSServer("127.0.0.1:0", records)
	if err != nil {
		t.Fatalf("unable to run test server: %v", err)
	}
	servers = append(servers, s)
	serverAddresses = append(serverAddresses, addr)

	return servers, serverAddresses
}

func startMockSemiBrokenDNSServerFleet(t *testing.T, records map[string][]string) ([]*dns.Server, []string) {
	servers := []*dns.Server{&dns.Server{}}
	serverAddresses := []string{"127.0.0.1:10000"}

	s, addr, err := startMockDNSServer("127.0.0.1:0", nil)
	if err != nil {
		t.Fatalf("unable to run test server: %v", err)
	}
	servers = append(servers, s)
	serverAddresses = append(serverAddresses, addr)

	s, addr, err = startMockDNSServer("127.0.0.1:0", records)
	if err != nil {
		t.Fatalf("unable to run test server: %v", err)
	}
	servers = append(servers, s)
	serverAddresses = append(serverAddresses, addr)

	s, addr, err = startMockDNSServer("127.0.0.1:0", records)
	if err != nil {
		t.Fatalf("unable to run test server: %v", err)
	}
	servers = append(servers, s)
	serverAddresses = append(serverAddresses, addr)

	return servers, serverAddresses
}

func stopMockDNSServerFleet(servers []*dns.Server) {
	for _, s := range servers {
		s.Shutdown()
	}
}

func Test_resolveARecord_noServer(t *testing.T) {
	dc := NewClient()
	_, err := dc.ResolveARecord("example.com.", []string{"127.0.0.1:65111"})
	if err == nil {
		t.Fatalf("resolveARecord should have returned an error")
	}
}

func Test_resolveARecord_empty(t *testing.T) {
	servers, serverAddresses := startMockDNSServerFleet(t, map[string][]string{"example.com.": []string{}})
	defer stopMockDNSServerFleet(servers)

	dc := NewClient()
	_, err := dc.ResolveARecord("example.com.", serverAddresses)
	if err != ErrEmptyAnswer {
		t.Fatalf("resolveARecord should have returned an empty answer error")
	}
}

func Test_resolveARecord_multipleDifferent(t *testing.T) {
	servers, serverAddresses := startMockDNSServerFleet(t, map[string][]string{"example.com.": []string{"1.1.1.1", "1.2.3.4"}})
	defer stopMockDNSServerFleet(servers)

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
	servers, serverAddresses := startMockDNSServerFleet(t, map[string][]string{"example.com.": []string{"1.1.1.1", "1.1.1.1"}})
	defer stopMockDNSServerFleet(servers)

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
	servers, serverAddresses := startMockSemiBrokenDNSServerFleet(t, map[string][]string{"example.com.": []string{"1.1.1.1", "1.1.1.1"}})
	defer stopMockDNSServerFleet(servers)

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
