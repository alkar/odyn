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

// Package odyn provides functionality to discover one's public IP address
// using a variety of services (providers).
//
// Public IP address Providers
//
// The services currently available to use through this package are ipify.org,
// ipinfo.io and opendns.com
//
// The easiest approach is to simply:
//  ip, err := IpifyProvider.Get()
//
// You can also use the HTTPProvider and DNSProvider to expand the
// functionality to further online services:
//
// For example, to retrieve it over HTTP:
//  p, err := NewHTTPProvider("http://myip.example.com")
//  ip, err := p.Get()
//
// The HTTPProvider and DNSProvider can be used to retrieve the public IP
// address from many different services since they're highly customisable.
// You can combine also them using ProviderSets:
//
//  ps, err := NewProviderSet(ProviderSetParallel, myHTTPProvider, myDNSProvider)
//  ip, err := ps.Get()
//
// See the documentation on NewProviderSet for more information.
//
// DNS Client
//
// To request for an A record from a set of nameservers:
//
//  c := NewDNSClient()
//  ip, err := c.ResolveA("test.example.com", []string{"8.8.8.8"})
//
// DNS Zone Providers
//
// DNS providers are tasked with updating A records:
//
//  p, err := NewRoute53Zone()
//  err := p.UpdateA("test.example.com", "example.com.", net.ParseIP("1.2.3.4"))
package odyn

import (
	"encoding/json"
	"fmt"
	"net"
)

const (
	// Version of the odyn package.
	Version = "master"
)

var (
	// IpifyProvider uses ipify.org to discover the public IP address.
	IpifyProvider, _ = NewHTTPProviderWithOptions(&HTTPProviderOptions{
		URL: "https://api.ipify.org",
		Headers: map[string]string{
			"User-Agent": fmt.Sprintf("odyn/%s", Version),
		},
	})

	// IPInfoProvider uses ipinfo.io to discover the public IP address.
	IPInfoProvider, _ = NewHTTPProviderWithOptions(&HTTPProviderOptions{
		URL:   "https://ipinfo.io",
		Parse: ipInfoParser,
		Headers: map[string]string{
			"Accept":     "application/json",
			"User-Agent": fmt.Sprintf("odyn/%s", Version),
		},
	})
	ipInfoParser = func(body []byte) (net.IP, error) {
		response := struct {
			IP           net.IP `json:"ip"`
			Hostname     string `json:"hostname"`
			City         string `json:"city"`
			Region       string `json:"region"`
			Country      string `json:"country"`
			Location     string `json:"loc"`
			Organisation string `json:"org"`
		}{}

		if err := json.Unmarshal(body, &response); err != nil {
			return nil, err
		}

		return response.IP, nil
	}

	// OpenDNSProvider uses OpenDNS's nameservers to discover the public IP
	// address.
	OpenDNSProvider, _ = NewDNSProvider("myip.opendns.com.", []string{
		"208.67.222.222:53", // resolver1.opendns.com
		"208.67.220.220:53", // resolver2.opendns.com
		"208.67.222.220:53", // resolver3.opendns.com
		"208.67.220.222:53", // resolver4.opendns.com
	})
)
