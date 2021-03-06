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
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	testCasesHTTP = []struct {
		resp string
		code int
		err  error
	}{
		{`1.2.3.4`, 200, nil},
		{`.2.3.4`, 200, ErrHTTPProviderCouldNotParseIP},
		{``, 500, ErrHTTPProviderInvalidResponseCode},
		{``, 200, ErrHTTPProviderCouldNotParseIP},
	}

	testHTTPProviderParser = func(body []byte) (net.IP, error) {
		response := struct {
			IPAddress net.IP `json:"ip"`
		}{}

		if err := json.Unmarshal(body, &response); err != nil {
			return nil, err
		}

		return response.IPAddress, nil
	}
)

func TestHTTPProvider_Get(t *testing.T) {
	var resp string
	var code int

	// create a mock http server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h, ok := r.Header["Header_key"]; ok && h[0] == "header_value" {
			w.WriteHeader(code)
			fmt.Fprintf(w, resp)
		}
	}))
	defer ts.Close()

	// mock the ip provider
	p, err := NewHTTPProviderWithOptions(&HTTPProviderOptions{
		URL: ts.URL,
		// Parse: testHTTPProviderParser,
		Headers: map[string]string{
			"Header_key": "header_value",
		},
	})
	if err != nil {
		t.Fatalf("NewHTTPProviderWithOptions returned unexpected error: %+v", err)
	}

	// test all cases
	for i, testCase := range testCasesHTTP {
		resp = testCase.resp
		code = testCase.code

		_, err := p.Get()

		if testCase.err == nil && err != nil {
			t.Errorf("HTTPProvider.Get returned unexpected error for case %02d: %+v", i, err)
			continue
		}

		if err == nil && testCase.err != nil {
			t.Errorf("HTTPProvider.Get did not return expected error for case %02d: %+v", i, err)
			continue
		}

		if err == nil && testCase.err == nil {
			continue
		}

		if err.Error() != testCase.err.Error() {
			t.Errorf("HTTPProvider.Get returned unexpected error for case %02d: %+v", i, err)
		}
	}
}

func TestHTTPProvider_Get_noConnectivity(t *testing.T) {
	// mock the ip provider
	p, err := NewHTTPProvider("https://127.0.0.1:64321")
	if err != nil {
		t.Fatalf("NewHTTPProviderWithOptions returned unexpected error: %+v", err)
	}

	_, err = p.Get()
	if err == nil {
		t.Errorf("HTTPProvider.Get did not return expected error")
	}

	if err.Error() != "Get https://127.0.0.1:64321: dial tcp 127.0.0.1:64321: getsockopt: connection refused" {
		t.Errorf("HTTPProvider.Get did not returned unexpected error: %+v", err)
	}
}

func TestNewHTTPProvider_error(t *testing.T) {
	if _, err := NewHTTPProvider(""); err != ErrHTTPProviderURLIsRequired {
		t.Errorf("NewHTTPProvider returned unexpected error: %+v", err)
	}
}

//
func TestNewHTTPProvider_invalidURL(t *testing.T) {
	if _, err := NewHTTPProvider(":"); err == nil {
		t.Errorf("NewHTTPProvider did not return an error")
	}
}
