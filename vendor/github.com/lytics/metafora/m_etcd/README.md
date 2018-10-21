metafora etcd client
====================

See [Documentation/etcd.md](../Documentation/etcd.md) for details.

Testing
-------

Testing the metafora etcd client requires that a new etcd instance be running.
The etcd instances should be reachable via the connection described by the 
connection string `localhost:5001,localhost:5002,localhost:5003` or a similar 
connection string should be exported as an environment variable `ETCDCTL_PEERS`.
The environemnt variable `ETCDTESTS` must be set, otherwise the tests will be
skipped.

An example of running the integration tests is given in the command line below:

```sh
ETCDTESTS=1 IP="127.0.0.1" ETCDCTL_PEERS="$IP:5001,$IP:5002,$IP:5003"  go test -v
```
