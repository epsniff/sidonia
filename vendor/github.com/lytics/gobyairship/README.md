# Go By Airship

[![GoDoc](https://godoc.org/github.com/lytics/gobyairship?status.svg)](https://godoc.org/github.com/lytics/gobyairship)
[![Build Status](https://travis-ci.org/lytics/gobyairship.svg?branch=master)](https://travis-ci.org/lytics/gobyairship)

Go client for Urban Airship

Currently only supports
[POSTs](https://godoc.org/github.com/lytics/gobyairship#Client.Post) and the
[Event Stream API](https://godoc.org/github.com/lytics/gobyairship/events).

## Testing

If you have Go 1.3 or later installed you can run tests with:

```sh
go get github.com/lytics/gobyairship
cd $GOPATH/src/github.com/lytics/gobyairship
go test ./...

# To run live API integration tests
UA_CREDS=<app key>:<access token> go test ./...
```
