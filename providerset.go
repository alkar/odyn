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

package odyn

import (
	"errors"
	"net"
	"sync"
)

const (
	// ProviderSetSerial sets the ProviderSet mode to serial.
	ProviderSetSerial = ProviderSetKind(1)

	// ProviderSetParallel sets the ProviderSet mode to parallel.
	ProviderSetParallel = ProviderSetKind(2)
)

var (
	// ErrProviderSetInvalidKind is returned when trying to create a new
	// ProviderSetor or querying using one with invalid kind set.
	ErrProviderSetInvalidKind = errors.New("invalid ProviderSet kind")

	// ErrProviderSetAllProvidersFailed is returned when all the providers in
	// the ProviderSet returned an error.
	ErrProviderSetAllProvidersFailed = errors.New("all the providers returned an error")

	// ErrProviderSetMultipleResults is returned when providers in a parallel
	// mode ProviderSet return different (conflicting) results.
	ErrProviderSetMultipleResults = errors.New("the providers returned multiple different results")
)

// ProviderSetKind represents the operational mode of a ProviderSet.
type ProviderSetKind int64

// ProviderSet combines multiple Providers to provide an easier way of
// handling results from multiple sources.
type ProviderSet struct {
	kind      ProviderSetKind
	providers []IPProvider
}

// NewProviderSet creates a new ProviderSet using the specified Providers.
// The kind argument dictates the method that the ProviderSet will use to
// extract the IP address:
//
// ProviderSetSerial will use the providers in sequence and return the first
// successful result, ignoring the rest of the Providers.
//
// ProviderSetParallel will query all the providers at once and ensure that
// they return the same result or return an error.
func NewProviderSet(kind ProviderSetKind, providers ...IPProvider) (*ProviderSet, error) {
	if kind != ProviderSetSerial && kind != ProviderSetParallel {
		return nil, ErrProviderSetInvalidKind
	}

	return &ProviderSet{
		kind:      kind,
		providers: providers,
	}, nil
}

// Get will use the providers to get the IP address.
func (p *ProviderSet) Get() (net.IP, error) {
	switch p.kind {
	case ProviderSetSerial:
		return p.getSerial()
	case ProviderSetParallel:
		return p.getParallel()
	default:
		return nil, ErrProviderSetInvalidKind
	}
}

func (p *ProviderSet) getSerial() (net.IP, error) {
	for _, provider := range p.providers {
		ip, err := provider.Get()
		if err == nil {
			return ip, err
		}
	}

	return nil, ErrProviderSetAllProvidersFailed
}

type providerSetParallelResult struct {
	IP    net.IP
	Error error
}

func providerSetParallelRun(wg *sync.WaitGroup, provider IPProvider, results chan providerSetParallelResult) {
	defer wg.Done()
	ip, err := provider.Get()
	results <- providerSetParallelResult{
		IP:    ip,
		Error: err,
	}
}

func (p *ProviderSet) getParallel() (net.IP, error) {
	results := make(chan providerSetParallelResult, len(p.providers))

	wg := &sync.WaitGroup{}
	wg.Add(len(p.providers))
	for _, provider := range p.providers {
		go providerSetParallelRun(wg, provider, results)
	}
	wg.Wait()
	close(results)

	ips := map[string]net.IP{}
	for r := range results {
		if r.Error == nil {
			ips[r.IP.String()] = r.IP
		}
	}

	if len(ips) != 1 {
		return nil, ErrProviderSetMultipleResults
	}

	var ip net.IP
	for _, v := range ips {
		ip = v
		break
	}

	return ip, nil
}
