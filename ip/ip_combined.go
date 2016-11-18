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
	"errors"
	"log"
	"net"
	"sync"
)

var (
	errCombinedBothFailed       = errors.New("Both providers returned an error")
	errCombinedDifferentResults = errors.New("The providers returned different results")
)

type ipProviderCombined struct {
	ipinfo  *ipProviderIPInfo
	opendns *ipProviderOpenDNS
}

func newIPProviderCombined() *ipProviderCombined {
	return &ipProviderCombined{
		ipinfo:  newIPProviderIPInfo(),
		opendns: newIPProviderOpenDNS(),
	}
}

func (p ipProviderCombined) Get() (net.IP, error) {
	var ipA, ipB net.IP
	var errA, errB error

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		ipA, errA = p.ipinfo.Get()
		if errA != nil {
			log.Printf("IPInfo provider return an error: %+v", errA)
		}
		wg.Done()
	}()

	go func() {
		ipB, errB = p.opendns.Get()
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
