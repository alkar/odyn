package main

import (
	"errors"
	"log"
	"net"
	"sync"
)

var (
	errCombinedBothFailed       = errors.New("Both providers returned an error")
	errCombinedDifferentResults = errors.New("The providers returned different results")
)

type ipProviderCombined struct{}

func (p ipProviderCombined) Get() (net.IP, error) {
	var ipA, ipB net.IP
	var errA, errB error

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		ipA, errA = (&ipProviderIPInfo{}).Get()
		if errA != nil {
			log.Printf("IPInfo provider return an error: %+v", errA)
		}
		wg.Done()
	}()

	go func() {
		ipB, errB = ipProviderOpenDNS{}.Get()
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
