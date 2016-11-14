package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"

	"github.com/miekg/dns"
)

const (
	ipInfoBaseURL         = "https://ipinfo.io"
	openDNSTargetHostname = "myip.opendns.com."
)

var (
	errIPInfoInvalidResponseCode = errors.New("IPInfo responded with a non-200 status code")
	errOpenDNSEmptyAnswer        = errors.New("OpenDNS answer was empty")
	errCombinedBothFailed        = errors.New("Both providers returned an error")
	errCombinedDifferentResults  = errors.New("The providers returned different results")

	openDNSNameservers = []string{
		"208.67.222.222:53", // resolver1.opendns.com
		"208.67.220.220:53", // resolver2.opendns.com
		"208.67.222.220:53", // resolver3.opendns.com
		"208.67.220.222:53", // resolver4.opendns.com
	}

	httpClientHeaders = map[string]string{
		"User-Agent": fmt.Sprintf("odyn/%s", odynVersion),
		"Accept":     "application/json",
	}
)

type providerPublicIP interface {
	Get() (string, error)
}

type providerIPInfo struct {
	IP           net.IP `json:"ip"`
	Hostname     string `json:"hostname"`
	City         string `json:"city"`
	Region       string `json:"region"`
	Country      string `json:"country"`
	Location     string `json:"loc"`
	Organisation string `json:"org"`
}

func (p *providerIPInfo) Get() (net.IP, error) {
	u, err := url.Parse(ipInfoBaseURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	for k, v := range httpClientHeaders {
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

type providerOpenDNS struct{}

func (p providerOpenDNS) Get() (net.IP, error) {
	c := dns.Client{}
	m := dns.Msg{}
	m.SetQuestion(openDNSTargetHostname, dns.TypeA)

	var retError error
	var retIP []net.IP

	for _, nameserver := range openDNSNameservers {
		r, _, err := c.Exchange(&m, nameserver)
		if err != nil {
			retError = err
			continue
		}

		if len(r.Answer) == 0 {
			retError = errOpenDNSEmptyAnswer
			continue
		}

		for _, ans := range r.Answer {
			ip := ans.(*dns.A).A

			exists := false
			for _, i := range retIP {
				if ip.Equal(i) {
					exists = true
					break
				}
			}

			if !exists {
				retIP = append(retIP, ip)
			}
		}

		retError = nil
		break
	}

	if len(retIP) != 1 {
		log.Printf("OpenDNS answer contained multiple entries, will only use the first: %+v", retIP)
	}

	return retIP[0], retError
}

type providerCombined struct{}

func (p providerCombined) Get() (net.IP, error) {
	var ipA, ipB net.IP
	var errA, errB error

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		ipA, errA = (&providerIPInfo{}).Get()
		if errA != nil {
			log.Printf("IPInfo provider return an error: %+v", errA)
		}
		wg.Done()
	}()

	go func() {
		ipB, errB = providerOpenDNS{}.Get()
		if errB != nil {
			log.Printf("OpenDNS provider return an error: %+v", errB)
		}
		wg.Done()
	}()

	wg.Wait()

	if errA != nil && errB != nil {
		return nil, errCombinedBothFailed
	}

	if errA != nil {
		return ipB, nil
	}

	if errB != nil {
		return ipA, nil
	}

	if !ipA.Equal(ipB) {
		return nil, errCombinedDifferentResults
	}

	return ipA, nil
}
