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

// Package dnstest provides some helpful functions to setup mock DNS
// nameservers for testing.
//
//  servers, serverAddresses, err := dnstest.StartMockDNSServerFleet(map[string][]string{
//    "test.example.com": []string{
//      "1.1.1.1",
//      "2.2.2.2",
//    },
//  })
//  defer dnstest.StopMockDNSServerFleet(servers)
//  if err != nil {
//    t.Fatalf("could not setup mock servers: %+v", err)
//  }
//
package dnstest

import (
	"net"
	"sync"
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

// StartMockDNSServerFleet starts four DNS nameservers that will respond based
// on the provided map of records and IP addresses.
func StartMockDNSServerFleet(records map[string][]string) ([]*dns.Server, []string, error) {
	servers := []*dns.Server{}
	serverAddresses := []string{}

	s, addr, err := startMockDNSServer("127.0.0.1:0", records)
	if err != nil {
		return nil, nil, err
	}
	servers = append(servers, s)
	serverAddresses = append(serverAddresses, addr)

	s, addr, err = startMockDNSServer("127.0.0.1:0", records)
	if err != nil {
		return nil, nil, err
	}
	servers = append(servers, s)
	serverAddresses = append(serverAddresses, addr)

	s, addr, err = startMockDNSServer("127.0.0.1:0", records)
	if err != nil {
		return nil, nil, err
	}
	servers = append(servers, s)
	serverAddresses = append(serverAddresses, addr)

	s, addr, err = startMockDNSServer("127.0.0.1:0", records)
	if err != nil {
		return nil, nil, err
	}
	servers = append(servers, s)
	serverAddresses = append(serverAddresses, addr)

	return servers, serverAddresses, nil
}

// StartMockSemiBrokenDNSServerFleet starts four DNS nameservers that will
// respond based on the provided map of records and IP addresses, however, the
// first nameserver is never started and is considered "broken".
func StartMockSemiBrokenDNSServerFleet(records map[string][]string) ([]*dns.Server, []string, error) {
	servers := []*dns.Server{&dns.Server{}}
	serverAddresses := []string{"127.0.0.1:10000"}

	s, addr, err := startMockDNSServer("127.0.0.1:0", nil)
	if err != nil {
		return nil, nil, err
	}
	servers = append(servers, s)
	serverAddresses = append(serverAddresses, addr)

	s, addr, err = startMockDNSServer("127.0.0.1:0", records)
	if err != nil {
		return nil, nil, err
	}
	servers = append(servers, s)
	serverAddresses = append(serverAddresses, addr)

	s, addr, err = startMockDNSServer("127.0.0.1:0", records)
	if err != nil {
		return nil, nil, err
	}
	servers = append(servers, s)
	serverAddresses = append(serverAddresses, addr)

	return servers, serverAddresses, nil
}

// StopMockDNSServerFleet calls Shutdown() and a list of dns.Servers.
func StopMockDNSServerFleet(servers []*dns.Server) {
	for _, s := range servers {
		s.Shutdown()
	}
}
