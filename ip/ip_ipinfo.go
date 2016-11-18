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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
)

const (
	ipInfoBaseURL = "https://ipinfo.io"
)

var (
	errIPInfoInvalidResponseCode = errors.New("IPInfo responded with a non-200 status code")

	ipInfoHTTPClientHeaders = map[string]string{
		"User-Agent": fmt.Sprintf("odyn/%s", odynVersion),
		"Accept":     "application/json",
	}
)

type ipInfoResponse struct {
	IP           net.IP `json:"ip"`
	Hostname     string `json:"hostname"`
	City         string `json:"city"`
	Region       string `json:"region"`
	Country      string `json:"country"`
	Location     string `json:"loc"`
	Organisation string `json:"org"`
}

type ipProviderIPInfo struct {
	client *http.Client
	url    *url.URL
}

func newIPProviderIPInfo() *ipProviderIPInfo {
	u, _ := url.Parse(ipInfoBaseURL)

	return &ipProviderIPInfo{
		client: &http.Client{},
		url:    u,
	}
}

func (p *ipProviderIPInfo) Get() (net.IP, error) {
	req, _ := http.NewRequest(http.MethodGet, p.url.String(), nil)

	for k, v := range ipInfoHTTPClientHeaders {
		req.Header.Set(k, v)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errIPInfoInvalidResponseCode
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	r := ipInfoResponse{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	return r.IP, nil
}
