package main

import (
	"errors"
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

type ipProviderOpenDNS struct{}

func (p ipProviderOpenDNS) Get() (net.IP, error) {
	return resolveARecord(openDNSTargetHostname, openDNSNameservers)
}
