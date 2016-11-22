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

package publicip

import (
	"encoding/json"
	"net"
)

// Provider is an interface for IP providers to implement.
type Provider interface {
	Get() (net.IP, error)
}

var (
	// IpifyProvider uses ipify.org to discover the public IP address.
	IpifyProvider, _ = NewHTTPProvider("https://api.ipify.org")

	// IPInfoProvider uses ipinfo.io to discover the public IP address.
	IPInfoProvider, _ = NewHTTPProviderWithOptions(&HTTPProviderOptions{
		URL: "https://ipinfo.io",
		Parse: func(body []byte) (net.IP, error) {
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
		},
		Headers: map[string]string{
			"Accept": "application/json",
		},
	})

	// OpenDNSProvider uses OpenDNS's nameservers to discover the public IP
	// address.
	OpenDNSProvider, _ = NewDNSProvider("myip.opendns.com.", []string{
		"208.67.222.222:53", // resolver1.opendns.com
		"208.67.220.220:53", // resolver2.opendns.com
		"208.67.222.220:53", // resolver3.opendns.com
		"208.67.220.222:53", // resolver4.opendns.com
	})
)
