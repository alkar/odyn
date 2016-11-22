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

// Package publicip provides functionality to discover one's public IP address
// using a variety of services (providers).
//
// Providers
//
// They are used to retrieve the public IP address from some web service.
//
// For example, to retrieve it over HTTP:
//  p, err := publicip.NewHTTPProvider("http://myip.example.com")
//  ip, err := IpifyProvider.Get()
//
// The HTTPProvider and DNSProvider can be used to retrieve the public IP
// address from many different services since they're highly customisable.
//
// ProviderSet
//
// You can also combine multiple providers:
//  ps, err := NewProviderSet(ProviderSetParallel, myHTTPProvider, myDNSProvider)
//  ip, err := ps.Get()
//
// See the documentation on NewProviderSet for more information.
package publicip
