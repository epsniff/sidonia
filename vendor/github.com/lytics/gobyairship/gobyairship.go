package gobyairship

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

var ErrTooManyRedirects = errors.New("too many redirects")

// Client is an Urban Airship API client. It handles authentication and
// provides helpers for making requests against the API.
type Client struct {
	// HTTPClient is the *http.Client to use when making requests. It defaults to
	// http.DefaultClient.
	HTTPClient *http.Client

	app_key      string
	access_token string
}

// NewClient creates a new Urban Airship API Client using the given App Key and
// Access Token.
func NewClient(app_key, access_token string) *Client {
	return &Client{
		HTTPClient:   http.DefaultClient,
		app_key:      app_key,
		access_token: access_token,
	}
}

// Post a request to the Urban Airship API with the Client's credentials. If
// body is non-nil it is marshaled to JSON and the appropriate headers are set.
//
// Extra headers an be added and will override any default values.
func (c *Client) Post(url string, body interface{}, extra http.Header) (*http.Response, error) {
	// Marshal body if it is non-nil
	var buf []byte
	if body != nil {
		var err error
		buf, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := c.newRequest("POST", url, buf)
	if err != nil {
		return nil, err
	}
	if extra != nil {
		for k, v := range extra {
			ck := http.CanonicalHeaderKey(k)
			req.Header[ck] = v
		}
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	// The Urban Airship API may respond with a 307 + Set-Cookie on POSTs which
	// is non-standard and therefore handled by this wrapper method instead of by
	// Go's http.Client. Give up after 10 redirects.
	try := 0
	const tries = 10
	for ; resp.StatusCode == http.StatusTemporaryRedirect && try < tries; try++ {
		// Cleanup body of redirect response so the connection will be reused
		ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		// POST to specified location (if one specified)
		loc, err := resp.Location()
		if err != nil && err != http.ErrNoLocation {
			return nil, err
		}
		if err == nil {
			// only set url if err != NoLocation
			url = loc.String()
		}

		req, err := c.newRequest("POST", url, buf)
		if err != nil {
			return nil, err
		}

		// Set the cookie token if it's sent
		if cookie := resp.Header.Get("Set-Cookie"); cookie != "" {
			req.Header.Add("Cookie", cookie)
		}
		resp, err = c.HTTPClient.Do(req)
		if err != nil {
			return nil, err
		}
	}
	if try == tries {
		// Exhausted retries; cleanup response and return an error
		ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, ErrTooManyRedirects
	}
	return resp, nil
}

// newRequest adds auth and accept headers to an Urban Airship API
// request. If buf is non-nil it is assumed to be JSON.
func (c *Client) newRequest(method, url string, buf []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-UA-Appkey", c.app_key)
	req.Header.Set("Authorization", "Bearer "+c.access_token)
	if len(buf) > 0 {
		req.Body = ioutil.NopCloser(bytes.NewReader(buf))
		req.Header.Set("Content-Type", "application/json")

		// Urban Airship APIs do not support chunked requests; set the Content-Length
		req.ContentLength = int64(len(buf))
	}
	return req, nil
}
