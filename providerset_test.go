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
	"testing"
)

var (
	errTestProvider    = errors.New("test error")
	testProviderOK     = &testProvider{IP: net.ParseIP("1.1.1.1"), Error: nil}
	testProviderBroken = &testProvider{IP: nil, Error: errTestProvider}
)

type testProvider struct {
	IP    net.IP
	Error error
}

func (p *testProvider) Get() (net.IP, error) {
	return p.IP, p.Error
}

func TestNewProviderSet_invalidKind(t *testing.T) {
	_, err := NewProviderSet(ProviderSetKind(3))
	if err != ErrProviderSetInvalidKind {
		t.Errorf("NewProviderSet returned unexpected error: %+v", err)
	}

	p, _ := NewProviderSet(ProviderSetKind(1))
	p.kind = ProviderSetKind(3)
	_, err = p.Get()
	if err != ErrProviderSetInvalidKind {
		t.Errorf("NewProviderSet returned unexpected error: %+v", err)
	}
}

func TestProviderSet_Serial_Get(t *testing.T) {
	ps, _ := NewProviderSet(ProviderSetSerial, testProviderBroken, testProviderOK)
	ip, err := ps.Get()
	if err != nil {
		t.Errorf("ProviderSet.Get returned unexpected error: %+v", err)
	}

	if !ip.Equal(testProviderOK.IP) {
		t.Errorf("ProviderSet.Get returned unexpected ip: %+v", ip)
	}
}

func TestProviderSet_Serial_Get_allFail(t *testing.T) {
	ps, _ := NewProviderSet(ProviderSetSerial, testProviderBroken)
	_, err := ps.Get()
	if err != ErrProviderSetAllProvidersFailed {
		t.Errorf("ProviderSet.Get returned unexpected error: %+v", err)
	}
}

func TestProviderSet_Parallel_Get(t *testing.T) {
	ps, _ := NewProviderSet(ProviderSetParallel, testProviderBroken, testProviderOK)
	ip, err := ps.Get()
	if err != nil {
		t.Errorf("ProviderSet.Get returned unexpected error: %+v", err)
	}

	if !ip.Equal(testProviderOK.IP) {
		t.Errorf("ProviderSet.Get returned unexpected ip: %+v", ip)
	}
}

func TestProviderSet_Parallel_Get_different(t *testing.T) {
	ps, _ := NewProviderSet(ProviderSetParallel, testProviderOK, &testProvider{IP: net.ParseIP("1.1.1.2"), Error: nil})
	_, err := ps.Get()
	if err != ErrProviderSetMultipleResults {
		t.Errorf("ProviderSet.Get returned unexpected error: %+v", err)
	}
}
