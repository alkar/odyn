package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var (
	testCases = []struct {
		resp string
		code int
		err  error
	}{
		{`{"ip": "1.2.3.4", "hostname": "", "city": "", "region": "", "country": "", "loc": "", "org": ""}`, 200, nil},
		{`{"ip": ".2.3.4", "hostname": "", "city": "", "region": "", "country": "", "loc": "", "org": ""}`, 200, errors.New("invalid IP address: .2.3.4")},
		{``, 500, errIPInfoInvalidResponseCode},
		{``, 200, errors.New("unexpected end of JSON input")},
	}
)

func TestIPProviderIPInfo_Get(t *testing.T) {
	var resp string
	var code int

	// create a mock http server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		fmt.Fprintf(w, resp)
	}))
	defer ts.Close()

	// mock the ip provider
	p := newIPProviderIPInfo()
	u, _ := url.Parse(ts.URL)
	p.url = u

	// test all cases
	for i, testCase := range testCases {
		resp = testCase.resp
		code = testCase.code

		_, err := p.Get()

		if testCase.err == nil && err != nil {
			t.Errorf("Was expecting no error but got: %+v", err)
			continue
		}

		if err == nil && testCase.err != nil {
			t.Error("Was expecting an error but got nil")
			continue
		}

		if err == nil && testCase.err == nil {
			continue
		}

		if err.Error() != testCase.err.Error() {
			t.Errorf("Got unexpected error for case %02d: %+v", i, err)
		}
	}
}
