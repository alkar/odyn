package odyn

import (
	"encoding/json"
	"fmt"
	"net"
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
