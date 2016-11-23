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

// Package odyn provides functionality to discover one's public IP address
// using a variety of services (providers).
//
// Public IP address Providers
//
// The services currently available to use through this package are ipify.org,
// ipinfo.io and opendns.com
//
// The easiest approach is to simply:
//  ip, err := IpifyProvider.Get()
//
// You can also use the HTTPProvider and DNSProvider to extend the
// functionality to further services, see the publicip package documentation.
//
// You can combine them using ProviderSets.
package odyn

const (
	Version = "master"
)
