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
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
)

var (
	// ErrRoute53ProviderNoHostedZoneFound is returned when the Route53Provider
	// fails to find the Route53 Hosted Zone.
	ErrRoute53ProviderNoHostedZoneFound = errors.New("could not find a Route53 hosted zone")

	// ErrRoute53ProviderWatchTimedOut is returned when the update method times
	// out waiting to confirm that the change has been applied.
	ErrRoute53ProviderWatchTimedOut = errors.New("timed out")

	defaultRoute53ProviderTTL           int64 = 60
	defaultRoute53ProviderWatchInterval       = 10 * time.Second
	defaultRoute53ProviderWatchTimeout        = 2 * time.Minute
)

type route53Zone struct {
	Name        string
	ID          string
	Nameservers []string
}

// Route53Provider is a DNS provider based on the Amazon Web Services Route53
// DNS service.
type Route53Provider struct {
	options *Route53ProviderOptions
}

// Route53ProviderOptions are used to alter the behaviour of the Route53Provider.
type Route53ProviderOptions struct {
	TTL            int64
	SessionOptions session.Options
	API            route53iface.Route53API
	WatchInterval  time.Duration
	WatchTimeout   time.Duration
}

// NewRoute53Provider returns a new instantiated Route53Provider with default
// options.
func NewRoute53Provider() (*Route53Provider, error) {
	return NewRoute53ProviderWithOptions(&Route53ProviderOptions{})
}

// NewRoute53ProviderWithOptions returns a new instantiated Route53Provider
// using the provided options.
func NewRoute53ProviderWithOptions(options *Route53ProviderOptions) (*Route53Provider, error) {
	if options.TTL == 0 {
		options.TTL = defaultRoute53ProviderTTL
	}

	if options.WatchInterval == 0 {
		options.WatchInterval = defaultRoute53ProviderWatchInterval
	}

	if options.WatchTimeout == 0 {
		options.WatchTimeout = defaultRoute53ProviderWatchTimeout
	}

	if options.API == nil {
		sess, err := session.NewSessionWithOptions(options.SessionOptions)
		if err != nil {
			return nil, err
		}

		options.API = route53.New(sess)
	}

	return &Route53Provider{options: options}, nil
}

// UpdateA will set the Route53 A Record in the specified zone to point to the
// provided IP address.
func (p *Route53Provider) UpdateA(recordName string, zoneName string, ip net.IP) error {
	zone, err := p.getZone(zoneName)
	if err != nil {
		return err
	}

	return p.updateRecord(recordName, zone.ID, ip)
}

// Nameservers returns the list of authoritative namservers for a DNS zone.
func (p *Route53Provider) Nameservers(zoneName string) ([]string, error) {
	zone, err := p.getZone(zoneName)
	if err != nil {
		return nil, err
	}

	return zone.Nameservers, nil
}

func (p *Route53Provider) updateRecord(recordName string, zoneID string, ip net.IP) error {
	resp, err := p.options.API.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String(route53.ChangeActionUpsert),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name:            aws.String(recordName),
						TTL:             aws.Int64(p.options.TTL),
						Type:            aws.String(route53.RRTypeA),
						ResourceRecords: []*route53.ResourceRecord{{Value: aws.String(ip.String())}},
					},
				},
			},
			Comment: aws.String("Managed by odyn"),
		},
		HostedZoneId: aws.String(zoneID),
	})
	if err != nil {
		return err
	}

	return p.waitForChange(*resp.ChangeInfo.Id)
}

func (p *Route53Provider) waitForChange(changeID string) error {
	timeout := time.NewTimer(p.options.WatchTimeout)
	tick := time.NewTicker(p.options.WatchInterval)
	defer func() {
		timeout.Stop()
		tick.Stop()
	}()

	var err error
	var change *route53.GetChangeOutput

	for {
		select {
		case <-tick.C:
			change, err = p.options.API.GetChange(&route53.GetChangeInput{Id: aws.String(changeID)})
			if err != nil {
				return err
			}

			if *change.ChangeInfo.Status == route53.ChangeStatusInsync {
				return nil
			}
		case <-timeout.C:
			return ErrRoute53ProviderWatchTimedOut
		}
	}
}

func (p *Route53Provider) getZone(name string) (*route53Zone, error) {
	zones, err := p.options.API.ListHostedZonesByName(&route53.ListHostedZonesByNameInput{
		DNSName:  aws.String(name),
		MaxItems: aws.String("1"),
	})
	if err != nil {
		return nil, err
	}

	if len(zones.HostedZones) == 0 || *zones.HostedZones[0].Name != name {
		return nil, ErrRoute53ProviderNoHostedZoneFound
	}

	zone, err := p.options.API.GetHostedZone(&route53.GetHostedZoneInput{
		Id: zones.HostedZones[0].Id,
	})
	if err != nil {
		return nil, err
	}

	ret := &route53Zone{
		ID:          *zone.HostedZone.Id,
		Nameservers: make([]string, len(zone.DelegationSet.NameServers)),
	}
	for i, ns := range zone.DelegationSet.NameServers {
		ret.Nameservers[i] = *ns
	}

	return ret, nil
}
