//   Copyright 2016 Lytics
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

/*
gowrapmx4j is a base library of types to assist UnMarshalling and Querying MX4J data.

MX4J is a very useful service which makes JMX data accessible via HTTP. Unfortunately little is done to
improve the data's representation and it is returned as dense raw XML via an API frought with perilous
query variables which are poorly documented.

The types and unmarshalling structures defined here have sorted out some of the XML saddness
returned from MX4J and makes it easier to operate on the data stuctures.

Why?
Java databases are still industry standard and there's a lot of mindshare built around them. Sadly their tools
can be very arcane or non-existant. This library is built specifically to help surface useful
information from Cassandra's MX4J endpoint to assist in debugging, monitoring, and management.

	A JSON API running in sidecar to MX4J is far more human readable, consumeable, and easier to
	engage with other services.

Basic API Primer:

Types* are the basic structs created to aid interaction/querying MX4J, unmarshall data from
XML endpoints.

The Registry is a concurrent safe map of MX4J data which is updated when queries are made.
This is to reduce the number of calls to MX4J if multiple goroutines want to access the data.

The Distill* API aids in cleaning up the data structures created from unmarshalling the
XML API. DistillAttribute and DistillAttributeTypes are the main functions which return
clean data structures for http endpoints.

Example:
  gowrapmx4j/cmd/cassandra_example/main.go
Showcases some ways to use features of gowrapmx4j
	Registry usage(registering endpoints, updating, and consumption)
	Wrapping the MX4J endpoint to update the registry
	Custom http JSON endpoints to expose JMX data cleanly!
*/
package gowrapmx4j
