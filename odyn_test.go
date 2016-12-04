package odyn

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
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
		w.WriteHeader(code)
		fmt.Fprintf(w, resp)
	}))
	defer ts.Close()

	// mock the ip provider
	p, _ := NewHTTPProviderWithOptions(&HTTPProviderOptions{
		URL:   ts.URL,
		Parse: ipInfoParser,
		Headers: map[string]string{
			"Accept": "application/json",
		},
	})

	// test all cases
	for i, testCase := range testCasesIPInfo {
		resp = testCase.resp
		code = testCase.code

		_, err := p.Get()

		if testCase.err == nil && err != nil {
			t.Errorf("IPInfo.Get() returned unexpected error: %+v", err)
			continue
		}

		if err == nil && testCase.err != nil {
			t.Error("IPInfo.Get() did not return an error as expected")
			continue
		}

		if err == nil && testCase.err == nil {
			continue
		}

		if err.Error() != testCase.err.Error() {
			t.Errorf("IPInfo.Get() returned unexpected error for case %02d: %+v", i, err)
		}
	}
}
