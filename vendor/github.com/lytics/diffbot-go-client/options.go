// Copyright 2014 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diffbot

import (
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Options holds the optional parameters for Diffbot client.
//
// See http://diffbot.com/products/automatic/
type Options struct {
	Fields                  string
	Timeout                 time.Duration
	Callback                string
	FrontpageAll            string
	Discussion              bool // discussion defaults to false
	ClassifierMode          string
	ClassifierStats         string
	BulkNotifyEmail         string
	BulkNotifyWebHook       string
	BulkRepeat              string
	BulkMaxRounds           string
	BulkPageProcessPattern  string
	CrawlMaxToCrawl         string
	CrawlMaxToProcess       string
	CrawlRestrictDomain     string
	CrawlNotifyEmail        string
	CrawlNotifyWebHook      string
	CrawlDelay              string
	CrawlRepeat             string
	CrawlOnlyProcessIfNew   string
	CrawlMaxRounds          string
	CrawlUrlPattern         string
	CrawlUrlRegexp          string
	CrawlUrlProcessPattern  string
	CrawlUrlProcessRegexp   string
	CrawlPageProcessPattern string
	CrawlMaxHops            string
	CrawlFormat             string
	CrawlType               string
	CrawlNumber             string
	BatchMethod             string
	BatchRelativeUrl        string
	CustomHeader            http.Header
}

// MethodParamString return string as the url params.
//
// If the Options is not empty, the return string begin with a '&'.
func (p *Options) MethodParamString(method string) url.Values {
	s := url.Values{}
	if p == nil || method == "" {
		return s
	}

	switch method {
	case "article", "image", "product":
		if p.Fields != "" {
			s.Add("fields", p.Fields)
		}
		if p.Timeout != 0 {
			timeout := strconv.FormatInt(int64(p.Timeout/time.Millisecond), 10)
			s.Add("timeout", timeout)
		}
		if p.Callback != "" {
			s.Add("callback", url.QueryEscape(p.Callback))
		}
		if !p.Discussion {
			s.Add("discussion", "false")
		}
		return s

	case "frontpage":
		if p.Timeout != 0 {
			timeout := strconv.FormatInt(int64(p.Timeout/time.Millisecond), 10)
			s.Add("timeout", timeout)
		}
		if p.FrontpageAll != "" {
			s.Add("all", p.FrontpageAll)
		}
		return s

	case "analyze":
		if p.ClassifierMode != "" {
			s.Add("mode", p.ClassifierMode)
		}
		if p.Fields != "" {
			s.Add("fields", p.Fields)
		}
		if p.ClassifierStats != "" {
			s.Add("stats", p.ClassifierStats)
		}
		if !p.Discussion {
			s.Add("discussion", "false")
		}
		return s

	case "bulk":
		if p.BulkNotifyEmail != "" {
			s.Add("notifyEmail", p.BulkNotifyEmail)
		}
		if p.BulkNotifyWebHook != "" {
			s.Add("notifyWebHook", p.BulkNotifyWebHook)
		}
		if p.BulkRepeat != "" {
			s.Add("repeat", p.BulkRepeat)
		}
		if p.BulkMaxRounds != "" {
			s.Add("maxRounds", p.BulkMaxRounds)
		}
		if p.BulkPageProcessPattern != "" {
			s.Add("pageProcessPattern", p.BulkPageProcessPattern)
		}
		return s

	case "crawl":
		if p.CrawlMaxToCrawl != "" {
			s.Add("maxToCrawl", p.CrawlMaxToCrawl)
		}
		if p.CrawlMaxToProcess != "" {
			s.Add("maxToProcess", p.CrawlMaxToProcess)
		}
		if p.CrawlRestrictDomain != "" {
			s.Add("restrictDomain", p.CrawlRestrictDomain)
		}
		if p.CrawlNotifyEmail != "" {
			s.Add("notifyEmail", p.CrawlNotifyEmail)
		}
		if p.CrawlNotifyWebHook != "" {
			s.Add("notifyWebHook", p.CrawlNotifyWebHook)
		}
		if p.CrawlDelay != "" {
			s.Add("crawlDelay", p.CrawlDelay)
		}
		if p.CrawlRepeat != "" {
			s.Add("repeat", p.CrawlRepeat)
		}
		if p.CrawlOnlyProcessIfNew != "" {
			s.Add("onlyProcessIfNew", p.CrawlOnlyProcessIfNew)
		}
		if p.CrawlMaxRounds != "" {
			s.Add("maxRounds", p.CrawlMaxRounds)
		}
		if p.CrawlUrlPattern != "" {
			s.Add("urlCrawlPattern", p.CrawlUrlPattern)
		}
		if p.CrawlUrlRegexp != "" {
			s.Add("urlCrawlRegEx", p.CrawlUrlRegexp)
		}
		if p.CrawlUrlProcessPattern != "" {
			s.Add("urlProcessPattern", p.CrawlUrlProcessPattern)
		}
		if p.CrawlUrlProcessRegexp != "" {
			s.Add("urlProcessRegEx", p.CrawlUrlProcessRegexp)
		}
		if p.CrawlPageProcessPattern != "" {
			s.Add("pageProcessPattern", p.CrawlPageProcessPattern)
		}
		if p.CrawlMaxHops != "" {
			s.Add("maxHops", p.CrawlMaxHops)
		}
		return s

	case "crawl/data":
		if p.CrawlFormat != "" {
			s.Add("format", p.CrawlFormat)
		}
		if p.CrawlType != "" {
			s.Add("type", p.CrawlType)
		}
		if p.CrawlNumber != "" {
			s.Add("num", p.CrawlNumber)
		}
		return s

	case "batch":
		if p.Timeout != 0 {
			timeout := strconv.FormatInt(int64(p.Timeout/time.Millisecond), 10)
			s.Add("timeout", timeout)
		}
		if p.BatchMethod != "" {
			s.Add("method", p.BatchMethod)
		}
		if p.BatchRelativeUrl != "" {
			s.Add("relative_urls", url.QueryEscape(p.BatchRelativeUrl))
		}
		return s

	default: // Custom APIs
		if p.Timeout != 0 {
			timeout := strconv.FormatInt(int64(p.Timeout/time.Millisecond), 10)
			s.Add("timeout", timeout)
		}
		if p.Callback != "" {
			s.Add("callback", url.QueryEscape(p.Callback))
		}
		return s
	}

	return s
}
