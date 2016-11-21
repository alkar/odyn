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

package dns

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
)

type mockRoute53API struct {
	route53iface.Route53API
	getZoneResp   *route53.GetHostedZoneOutput
	getZoneErr    error
	getChangeResp *route53.GetChangeOutput
	getChangeErr  error
	listZonesResp *route53.ListHostedZonesByNameOutput
	listZonesErr  error
	changeRRResp  *route53.ChangeResourceRecordSetsOutput
	changeRRErr   error
}

func (m mockRoute53API) ListHostedZonesByName(in *route53.ListHostedZonesByNameInput) (*route53.ListHostedZonesByNameOutput, error) {
	return m.listZonesResp, m.listZonesErr
}

func (m mockRoute53API) GetHostedZone(in *route53.GetHostedZoneInput) (*route53.GetHostedZoneOutput, error) {
	return m.getZoneResp, m.getZoneErr
}

func (m mockRoute53API) ChangeResourceRecordSets(in *route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error) {
	return m.changeRRResp, m.changeRRErr
}

func (m mockRoute53API) GetChange(in *route53.GetChangeInput) (*route53.GetChangeOutput, error) {
	return m.getChangeResp, m.getChangeErr
}

func TestRoute53Provider_defaults(t *testing.T) {
	p, _ := NewRoute53Provider()

	if p.options.TTL != defaultRoute53ProviderTTL {
		t.Errorf("NewRoute53Provider default TTL is not what was expected: %+v", p.options.TTL)
	}

	if p.options.WatchInterval != defaultRoute53ProviderWatchInterval {
		t.Errorf("NewRoute53Provider default WatchInterval is not what was expected: %+v", p.options.WatchInterval)
	}

	if p.options.WatchTimeout != defaultRoute53ProviderWatchTimeout {
		t.Errorf("NewRoute53Provider default WatchTimeout is not what was expected: %+v", p.options.WatchTimeout)
	}
}

var (
	errTestRoute53ProviderMock = errors.New("test error")

	testRoute53ProviderListZonesOK = &route53.ListHostedZonesByNameOutput{
		DNSName: aws.String("example.com."),
		HostedZones: []*route53.HostedZone{
			&route53.HostedZone{
				ResourceRecordSetCount: aws.Int64(1),
				CallerReference:        aws.String(""),
				Config: &route53.HostedZoneConfig{
					Comment:     aws.String(""),
					PrivateZone: aws.Bool(false),
				},
				Id:   aws.String("/hostedzone/XXXXXXXXXXXXXX"),
				Name: aws.String("example.com."),
			},
		},
		NextHostedZoneId: aws.String(""),
		MaxItems:         aws.String("1"),
		NextDNSName:      aws.String(""),
		IsTruncated:      aws.Bool(false),
	}

	testRoute53ProviderGetZoneOK = &route53.GetHostedZoneOutput{
		HostedZone: &route53.HostedZone{
			ResourceRecordSetCount: aws.Int64(1),
			CallerReference:        aws.String(""),
			Config: &route53.HostedZoneConfig{
				Comment:     aws.String(""),
				PrivateZone: aws.Bool(false),
			},
			Id:   aws.String("/hostedzone/XXXXXXXXXXXXXX"),
			Name: aws.String("example.com."),
		},
		DelegationSet: &route53.DelegationSet{
			NameServers: []*string{
				aws.String("0.ns.example.com"),
				aws.String("1.ns.example.com"),
				aws.String("2.ns.example.com"),
				aws.String("3.ns.example.com"),
			},
			CallerReference: aws.String(""),
			Id:              aws.String("/delegationset/XXXXXXXXXXXXXX"),
		},
	}

	testRoute53ProviderChangeRROK = &route53.ChangeResourceRecordSetsOutput{
		ChangeInfo: &route53.ChangeInfo{
			Comment:     aws.String(""),
			Id:          aws.String("123456789"),
			Status:      aws.String(route53.ChangeStatusPending),
			SubmittedAt: aws.Time(time.Now()),
		},
	}

	testRoute53ProviderGetChangeOK = &route53.GetChangeOutput{
		ChangeInfo: &route53.ChangeInfo{
			Comment:     aws.String(""),
			Id:          aws.String("123456789"),
			Status:      aws.String(route53.ChangeStatusInsync),
			SubmittedAt: aws.Time(time.Now()),
		},
	}

	testRoute53ProviderGetChangePending = &route53.GetChangeOutput{
		ChangeInfo: &route53.ChangeInfo{
			Comment:     aws.String(""),
			Id:          aws.String("123456789"),
			Status:      aws.String(route53.ChangeStatusPending),
			SubmittedAt: aws.Time(time.Now()),
		},
	}
)

func TestRoute53Provider_Update(t *testing.T) {
	testCases := []struct {
		listZonesErr      error
		listZonesResponse *route53.ListHostedZonesByNameOutput
		getZoneErr        error
		getZoneResponse   *route53.GetHostedZoneOutput

		changeRRErr       error
		changeRRResponse  *route53.ChangeResourceRecordSetsOutput
		getChangeErr      error
		getChangeResponse *route53.GetChangeOutput

		recordName string
		zoneName   string
		ip         net.IP

		expectedErr error
	}{
		{ // error in list zones
			errTestRoute53ProviderMock,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			"test.example.com",
			"example.com.",
			net.ParseIP("1.1.1.1"),
			errTestRoute53ProviderMock,
		},
		{ // zone not found
			nil,
			testRoute53ProviderListZonesOK,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			"",
			"test.example.com",
			net.ParseIP("1.1.1.1"),
			ErrRoute53ProviderNoHostedZoneFound,
		},
		{ // error in get zone
			nil,
			testRoute53ProviderListZonesOK,
			errTestRoute53ProviderMock,
			nil,
			nil,
			nil,
			nil,
			nil,
			"test.example.com",
			"example.com.",
			net.ParseIP("1.1.1.1"),
			errTestRoute53ProviderMock,
		},
		{ // error in change request
			nil,
			testRoute53ProviderListZonesOK,
			nil,
			testRoute53ProviderGetZoneOK,
			errTestRoute53ProviderMock,
			nil,
			nil,
			nil,
			"test.example.com",
			"example.com.",
			net.ParseIP("1.1.1.1"),
			errTestRoute53ProviderMock,
		},
		{ // error in get change request
			nil,
			testRoute53ProviderListZonesOK,
			nil,
			testRoute53ProviderGetZoneOK,
			nil,
			testRoute53ProviderChangeRROK,
			errTestRoute53ProviderMock,
			nil,
			"test.example.com",
			"example.com.",
			net.ParseIP("1.1.1.1"),
			errTestRoute53ProviderMock,
		},
		{ // timeout in get change
			nil,
			testRoute53ProviderListZonesOK,
			nil,
			testRoute53ProviderGetZoneOK,
			nil,
			testRoute53ProviderChangeRROK,
			nil,
			testRoute53ProviderGetChangePending,
			"test.example.com",
			"example.com.",
			net.ParseIP("1.1.1.1"),
			ErrRoute53ProviderWatchTimedOut,
		},
		{ // works end to end
			nil,
			testRoute53ProviderListZonesOK,
			nil,
			testRoute53ProviderGetZoneOK,
			nil,
			testRoute53ProviderChangeRROK,
			nil,
			testRoute53ProviderGetChangeOK,
			"test.example.com",
			"example.com.",
			net.ParseIP("1.1.1.1"),
			nil,
		},
	}

	for _, tc := range testCases {
		p, _ := NewRoute53ProviderWithOptions(&Route53ProviderOptions{
			API: &mockRoute53API{
				getZoneResp: tc.getZoneResponse,
				getZoneErr:  tc.getZoneErr,

				getChangeResp: tc.getChangeResponse,
				getChangeErr:  tc.getChangeErr,

				listZonesResp: tc.listZonesResponse,
				listZonesErr:  tc.listZonesErr,
				changeRRResp:  tc.changeRRResponse,
				changeRRErr:   tc.changeRRErr,
			},
			WatchInterval: 100 * time.Millisecond,
			WatchTimeout:  time.Second,
		})

		err := p.Update(tc.recordName, tc.zoneName, tc.ip)
		if err != tc.expectedErr {
			t.Errorf("Route53Provider.Update returned unexpected error: %+v", err)
		}
	}
}
