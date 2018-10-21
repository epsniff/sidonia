Go Wrap MX4J
------------

[![GoDoc](https://godoc.org/github.com/lytics/gowrapmx4?status.svg)](https://godoc.org/github.com/lytics/gowrapmx4j)

gowrapmx4j is a base library to assist interacting with the MX4J service.

MX4J is a very useful service which makes JMX data accessible via HTTP. Unfortunately little is done to
improve the data's representation and it is returned as dense raw XML via an API frought with perilous
query variables which are poorly documented.

The types and unmarshalling structures defined here have sorted out some of the XML saddness
returned from MX4J and makes it easier to operate on the data stuctures.

### Why
Java databases are still industry standard and there's a lot of mindshare built around them. Sadly their tools
can be very arcane or non-existant. This library is built specifically to help surface useful
information from Cassandra's MX4J endpoint to assist in debugging, monitoring, and management.

This library was built against [Cassandra 2.1.10](http://cassandra.apache.org/download/) and MX4J `3.0.2`. The tooling should work against other Java services which run with MX4J. 

A JSON API running in sidecar is far more human readable, consume, and engage with other services. 

### Basic API Primer

Types `types.go` are the basic structs created to aid interaction/querying MX4J, unmarshal data from XML endpoints.

The Registry `registry.go` is a concurrent safe map of MX4J data which is updated when queries are made.
This is to reduce the number of calls to MX4J if multiple goroutines want to access the data.

The Distill `distill.go` API aids in cleaning up the data structures created from unmarshalling the
XML API. DistillAttribute and DistillAttributeTypes are the main functions which return
clean data structures for http endpoints.

### Where to start
An example web service which operates in a sidecar pattern to the Cassandra/MX4J services provides nice
example usage: `gowrapmx4j/cmd/cassandra_example/main.go`

Showcases gowrapmx4j components:  

* Data structures to unmarshal XML data into Go structs
* Registry usage(registering endpoints, updating, and consumption)
* Wrapping the MX4J endpoint for easy polling to update the registry
* Custom http JSON endpoints to expose JMX data cleanly!

### TODO:
* Metrics examples; currently the metric functionality isn't showcased however it is possible and creating useful keyspace level metrics for production environments. 

## Contact
[Josh Roppo](https://github.com/Ropes) is the primary developer on the project.

Critique, ideas, and PRs welcome!

### External Requirements
[Logrus](https://github.com/Sirupsen/logrus); for nice log handling.

If using the [Go Vendor Experiment](https://medium.com/@freeformz/go-1-5-s-vendor-experiment-fd3e830f52c3#.fq1ap96hb) everything should just work(yay GO 1.6!). Otherwise you might need to `go get -u github.com/Sirupsen/logrus` to make it available in your GOPATH.

