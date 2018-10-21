// Copyright 2014 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diffbot

import (
	"encoding/json"
)

// Error represents an Diffbot APIs returns error.
//
// When issues arise, Diffbot APIs return the following fields in a JSON response.
//
// Simple Error code:
//
//	{
//		"error": "Could not download page (404)",
//		"errorCode": 404
//	}
//
// Possible errors returned:
//
//	+------+-----------------------------------------------------------------------------------------------------+
//	| CODE | DESCRIPTION                                                                                         |
//	+------+-----------------------------------------------------------------------------------------------------+
//	| 401  | Unauthorized token                                                                                  |
//	| 404  | Requested page not found                                                                            |
//	| 429  | Your token has exceeded the allowed number of calls, or has otherwise been throttled for API abuse. |
//	| 500  | Error processing the page. Specific information will be returned in the JSON response.              |
//	+------+-----------------------------------------------------------------------------------------------------+
//
type Error struct {
	ErrCode    int    `json:"errorCode"` // Description of the error
	ErrMessage string `json:"error"`     // Error code per the chart below
	RawString  string `json:"-"`         // Raw json format error string
}

// ParseJson parses the JSON-encoded error data.
func (p *Error) ParseJson(s string) error {
	if err := json.Unmarshal([]byte(s), p); err != nil {
		return err
	}
	p.RawString = s
	return nil
}

func (p *Error) Error() string {
	d, _ := json.Marshal(p)
	return string(d)
}
