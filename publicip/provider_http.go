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

package publicip

import (
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
)

var (
	// ErrHTTPProviderInvalidResponseCode is returned when the HTTP service
	// responds with a non 200 HTTP code.
	ErrHTTPProviderInvalidResponseCode = errors.New("provider responded with a non-200 status code")

	// ErrHTTPProviderCouldNotParseIP is returned when the provider is unable
	// to parse the IP address from the response body.
	ErrHTTPProviderCouldNotParseIP = errors.New("provider could not parse IP address from the response")

	// ErrHTTPProviderURLIsRequired is returned when trying to create an
	// HTTPProvider with a nil URL address.
	ErrHTTPProviderURLIsRequired = errors.New("the URL option is required")

	defaultHTTPProviderClient = &http.Client{}

	defaultHTTPProviderRequester = func(options *HTTPProviderOptions) ([]byte, error) {
		req, err := http.NewRequest(http.MethodGet, options.URL, nil)
		if err != nil {
			return nil, err
		}

		for k, v := range options.Headers {
			req.Header.Set(k, v)
		}

		resp, err := options.Client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, ErrHTTPProviderInvalidResponseCode
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		defer func() {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}()

		return body, nil
	}

	defaultHTTPProviderParser = func(body []byte) (net.IP, error) {
		ip := net.ParseIP(string(body))
		if ip == nil {
			return nil, ErrHTTPProviderCouldNotParseIP
		}

		return ip, nil
	}
)

// HTTPProvider is used to retrieve the IP address from any HTTP service.
type HTTPProvider struct {
	options *HTTPProviderOptions
}

// HTTPProviderOptions are used to alter the behaviour of the JSON IP Provider.
type HTTPProviderOptions struct {
	// Function to send the HTTP request to the service and return the response
	// body.
	Request HTTPProviderRequester

	// Function to parse the response body and return an IP address.
	Parse HTTPProviderResponseParser

	// HTTP Client used to send the GET request.
	Client *http.Client

	// HTTP Headers to set on the request
	Headers map[string]string

	// URL endpoint of the service.
	URL string
}

// HTTPProviderRequester is tasked with sending an HTTP request to the service
// ander returning the body of the response if the request was successful.
type HTTPProviderRequester func(options *HTTPProviderOptions) ([]byte, error)

// HTTPProviderResponseParser is tasked with parsing the HTTP response body and
// returning an IP address.
type HTTPProviderResponseParser func(body []byte) (net.IP, error)

// NewHTTPProvider returns a basic HTTPProvider that can parse plaintext
// responses from a URL.
func NewHTTPProvider(u string) (*HTTPProvider, error) {
	return NewHTTPProviderWithOptions(&HTTPProviderOptions{URL: u})
}

// NewHTTPProviderWithOptions allows you to specify the HTTPProviderOptions
// and completely customise the behaviour.
func NewHTTPProviderWithOptions(options *HTTPProviderOptions) (*HTTPProvider, error) {
	if options.URL == "" {
		return nil, ErrHTTPProviderURLIsRequired
	}

	if _, err := url.Parse(options.URL); err != nil {
		return nil, err
	}

	if options.Client == nil {
		options.Client = defaultHTTPProviderClient
	}

	if options.Parse == nil {
		options.Parse = defaultHTTPProviderParser
	}

	if options.Request == nil {
		options.Request = defaultHTTPProviderRequester
	}

	return &HTTPProvider{options: options}, nil
}

// Get will discover the public IP address using the HTTP service defined in
// the options of the HTTPProvider.
func (p *HTTPProvider) Get() (net.IP, error) {
	body, err := p.options.Request(p.options)
	if err != nil {
		return nil, err
	}

	return p.options.Parse(body)
}
