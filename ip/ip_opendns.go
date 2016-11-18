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
