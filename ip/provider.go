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
	"encoding/json"
	"net"
	"net/url"
)

var (
	ipinfoURL, _ = url.Parse("https://ipinfo.io")

	// IPInfoProvider uses ipinfo.io to discover the public IP address.
	IPInfoProvider, _ = NewHTTPProviderWithOptions(&HTTPProviderOptions{
		URL: ipinfoURL,
		Parse: func(body []byte) (net.IP, error) {
			response := struct {
				IPAddress    net.IP `json:"ip"`
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

			return response.IPAddress, nil
		},
	})
)

