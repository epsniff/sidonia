// Copyright 2014 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diffbot

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	DefaultServer = `http://api.diffbot.com/v3`
)

type Request struct {
	PageUrl         string   `json:"pageUrl"`
	ResolvedPageUrl string   `json:"resolvedPageUrl,omitempty"`
	API             string   `json:"api"`
	Options         []string `json:"options"`
	Fallback        string   `json:"fallback,omitempty"`
	Fields          string   `json:"fields,omitempty"`
	Version         int      `json:"version"`
}

// Diffbot uses computer vision, natural language processing
// and machine learning to automatically recognize
// and structure specific page-types.
func Diffbot(client *http.Client, method, token, url string, opt *Options) (body []byte, err error) {
	return DiffbotServer(client, DefaultServer, method, token, url, opt)
}

// DiffbotServer like Diffbot function, but support custom server.
func DiffbotServer(client *http.Client, server, method, token, url string, opt *Options) (body []byte, err error) {
	req, err := http.NewRequest("GET", makeRequestUrl(server, method, token, url, opt), nil)
	if err != nil {
		return nil, err
	}
	if opt != nil {
		if opt.CustomHeader != nil {
			req.Header.Add("X-Forward-User-Agent", opt.CustomHeader.Get("User-Agent"))
			req.Header.Add("X-Forward-Referer", opt.CustomHeader.Get("Forward-Referer"))
			req.Header.Add("X-Forward-Cookie", opt.CustomHeader.Get("Cookie"))
			req.Header.Add("X-Forward-Accept-Language", opt.CustomHeader.Get("Accept-Language"))
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if len(body) != 0 {
		var apiError Error
		if err = apiError.ParseJson(string(body)); err != nil {
			err = &Error{
				ErrCode:    resp.StatusCode,
				ErrMessage: string(body),
			}
		} else {
			if apiError.ErrCode != 0 {
				err = &apiError
				return
			}
		}
	} else {
		err = &Error{
			ErrCode:    resp.StatusCode,
			ErrMessage: resp.Status,
		}
		return
	}

	return
}

func makeRequestUrl(server, method, token, webUrl string, opt *Options) string {
	query := opt.MethodParamString(method)
	query.Add("token", token)
	query.Add("url", webUrl)
	return fmt.Sprintf("%s/%s?%s", server, method, query.Encode())
}
