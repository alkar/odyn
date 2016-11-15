package main

import (
	"encoding/json"
	"errors"
	"fmt"
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

type ipProviderIPInfo struct {
	IP           net.IP `json:"ip"`
	Hostname     string `json:"hostname"`
	City         string `json:"city"`
	Region       string `json:"region"`
	Country      string `json:"country"`
	Location     string `json:"loc"`
	Organisation string `json:"org"`
}

func (p *ipProviderIPInfo) Get() (net.IP, error) {
	u, err := url.Parse(ipInfoBaseURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	for k, v := range ipInfoHTTPClientHeaders {
		req.Header.Set(k, v)
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errIPInfoInvalidResponseCode
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, p)
	if err != nil {
		return nil, err
	}

	return p.IP, nil
}
