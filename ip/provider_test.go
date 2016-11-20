package ip

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var (
	testCasesIPInfo = []struct {
		resp string
		code int
		err  error
	}{
		{`{"ip": "1.2.3.4", "hostname": "", "city": "", "region": "", "country": "", "loc": "", "org": ""}`, 200, nil},
		{`{"ip": ".2.3.4", "hostname": "", "city": "", "region": "", "country": "", "loc": "", "org": ""}`, 200, errors.New("invalid IP address: .2.3.4")},
		{``, 500, ErrHTTPProviderInvalidResponseCode},
		{``, 200, errors.New("unexpected end of JSON input")},
	}
)

func TestIPInfoProvider_Get(t *testing.T) {
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
	u, _ := url.Parse(ts.URL)
	originalURL := IPInfoProvider.options.URL
	IPInfoProvider.options.URL = u
	defer func() {
		IPInfoProvider.options.URL = originalURL
	}()

	// test all cases
	for i, testCase := range testCasesIPInfo {
		resp = testCase.resp
		code = testCase.code

		_, err := IPInfoProvider.Get()

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
