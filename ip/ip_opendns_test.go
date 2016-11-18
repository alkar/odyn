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
