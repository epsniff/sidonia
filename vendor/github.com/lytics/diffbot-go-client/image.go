// Copyright 2014 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diffbot

import (
	"encoding/json"
	"net/http"
)

// See http://diffbot.com/dev/docs/image/
type Image struct {
	Type            string `json:"type"`
	Url             string `json:"url"`
	Title           string `json:"title,omitempty"`
	NaturalHeight   int    `json:"naturalHeight"`
	NaturalWidth    int    `json:"naturalWidth"`
	HumanLanguage   string `json:"humanLanguage,omitempty"`
	AnchorUrl       string `json:"anchorUrl,omitempty"`
	PageUrl         string `json:"pageUrl,omitempty"`
	ResolvedPageUrl string `json:"resolvedPageUrl,omitempty"`
	XPath           string `json:"xpath,omitempty"`
	DiffbotUri      string `json:"diffbotUri"`

	// optional image fields
	DisplayHeight int                      `json:"displayHeight,omitempty"`
	DisplayWidth  int                      `json:"displayWidth,omitempty"`
	Mentions      []string                 `json:"mentions,omitempty"`
	Ocr           string                   `json:"ocr,omitempty"`
	Faces         []map[string]interface{} `json:"faces,omitempty"`

	// optional fields
	Breadcrumb  []*breadcrumb          `json:"breadcrumb,omitempty"`
	Links       []string               `json:"links,omitempty"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
	QueryString string                 `json:"querystring,omitempty"`
}

// The Image API identifies the primary image(s) of a submitted web page and returns
// comprehensive information and metadata for each image.
//
// Request
//
// To use the Image API, perform a HTTP GET request on the following endpoint:
//
//  http://api.diffbot.com/v3/image
//
// Provide the following arguments:
//
//	+----------+-------------------------------------------------------------------------+
//	| ARGUMENT | DESCRIPTION                                                             |
//	+----------+-------------------------------------------------------------------------+
//	| token    | Developer token                                                         |
//	| url      | Web page URL of the image to process (URL encoded)                      |
//	+----------+-------------------------------------------------------------------------+
//	| Optional arguments                                                                 |
//	+----------+-------------------------------------------------------------------------+
//	| fields   | Used to specify optional fields to be returned by the Image API. See    |
//  |          | the Fields section below.                                               |
//	| timeout  | Sets a value in milliseconds to wait for the retrieval/fetch of content |
//  |          | from the requested URL. The default timeout for the third-party response|
//  |          | is 30 seconds (30000).                                                  |
//	| callback | Use for jsonp requests. Needed for cross-domain ajax.                   |
//	+----------+-------------------------------------------------------------------------+
//
// The fields argument
// Use the fields argument to return optional fields in the JSON response. The default fields
// will always be returned. For nested arrays, use parentheses to retrieve specific fields,
// or * to return all sub-fields.
//
// For example, to return meta (in addition to the default fields), your &fields argument
// would be:
//
//  &fields=meta
//
// Response
//
// The Image API returns data in JSON format.
//
// Each V3 response includes a request object (which returns request-specific metadata), and
// an objects array, which will include the extracted information for all images on a submitted
// page.
//
// Objects in the Image API's objects array will include the following fields:
//
//	+---------------+------------------------------------------------------------------------+
//	| FIELD         | DESCRIPTION                                                            |
//	+---------------+------------------------------------------------------------------------+
//	| type          | Type of object (always image)                                          |
//	| url           | Direct link to image file.                                             |
//	| title         | Title or caption of the image, if available.                           |
//	| naturalHeight | Raw image height, in pixels.                                           |
//	| naturalWidth  | Raw image width, in pixels.                                            |
//	| humanLanguage | Returns the (spoken/human) language of the submitted page, using       |
//	|               | two-letter ISO 639-1 nomenclature.                                     |
//	| anchorUrl     | If the image is hyperlinked, returns the destination URL.              |
//	| pageUrl       | URL of submitted page / page from which the image is extracted.        |
//	|resolvedPageUrl| Returned if the pageUrl redirects to another URL.                      |
//	| xpath         | XPath expression identifying the image node.                           |
//	| diffbotUri    | Unique object ID. The diffbotUri is generated from the values of       |
//	|               | various Image fields and uniquely identifies the object. This can be   |
//	|               | used for deduplication.                                                |
//	+----------------------------------------------------------------------------------------+
//	| Optional fields, available using fields= argument                                      |
//	+----------------------------------------------------------------------------------------+
//	| displayHeight | Height of image as presented in the browser (and as sized via          |
//	|               | browser/CSS, if resized).                                              |
//	| displayWidth  | Width of image as presented in the browser (and as sized via           |
//	|               | browser/CSS, if resized).                                              |
//	| links         | Returns a top-level object (links) containing all hyperlinks found on  |
//	|               | the page.                                                              |
//	| meta          | Comma-separated list of image-embedded metadata (e.g., EXIF, XMP, ICC  |
//	|               | Profile), if available within the image file.                          |
//	| querystring   | Returns any key/value pairs present in the URL querystring. Items      |
//	|               | without a discrete value will be returned as true.                     |
//	| breadcrumb    | Returns a top-level array (breadcrumb) of URLs and link text from page |
//	|               | breadcrumbs.                                                           |
//	+----------------------------------------------------------------------------------------+
//	| The following fields are in an early beta stage:                                       |
//	+----------------------------------------------------------------------------------------+
//	| mentions      | Array of articles upon which the same or similar image may be found.   |
//	| ocr           | If text is identified within the image, we will attempt to recognize   |
//	|               | the text string.                                                       |
//	| faces         | The x, y, height and width of coordinates of human faces. Returns null |
//	|               | if no faces are found.                                                 |
//	+---------------+------------------------------------------------------------------------+
//
// Example Response
//
// This is a simple response:
//
//  {
//    "request": {
//      "pageUrl": "http://www.diffbot.com/products",
//      "api": "image",
//      "options": [],
//      "fields": "",
//      "version": 3
//    },
//    {
//    "objects": [
//      {
//        "title": "Diffy, climbing a mountain",
//        "naturalHeight": 1158,
//        "diffbotUri": "image|3|-1897071612",
//        "pageUrl": "http://www.diffbot.com/products",
//        "humanLanguage": "en",
//        "naturalWidth": 950,
//        "date": "Oct 19, 2013",
//        "type": "image",
//        "url": "http://www.diffbot.com/images/image_diffy_sample.png",
//        "xpath": "/HTML/BODY/DIV[@class='main']/DIV[@id='primaryImage']/IMG"
//      },
//      {
//        "title": "Diffy atop said mountain",
//        "naturalHeight": 1120,
//        "diffbotUri": "image|3|-1221792290",
//        "pageUrl": "http://www.diffbot.com/products",
//        "humanLanguage": "en",
//        "naturalWidth": 920,
//        "anchorUrl": "http://www.diffbot.com",
//        "date": "Oct 21, 2013",
//        "type": "image",
//        "url": "http://www.diffbot.com/images/image_atopmountain_sample.png",
//        "xpath": "/HTML/BODY/DIV[@class='main']/DIV[@id='secondaryImage']/A/IMG"
//      },
//    ],
//  }
//
// Authentication
//
// You can supply Diffbot with basic authentication credentials or custom HTTP headers (see below)
// to access intranet pages or other sites that require a login.
//
// Basic Authentication
// To access pages that require a login/password (using basic access authentication), include the
// username and password in your url parameter, e.g.: url=http%3A%2F%2FUSERNAME:PASSWORD@www.diffbot.com.
//
// Custom HTTP Headers
//
// You can supply Diffbot APIs with custom values for the user-agent, referer, cookie, or accept-language
// values in the HTTP request. These will be used in place of the Diffbot default values.
//
// To provide custom headers, pass in the following values in your own headers when calling the Diffbot API:
//
//	+----------------------+-----------------------------------------------------------------------+
//	| HEADER               | DESCRIPTION                                                           |
//	+----------------------+-----------------------------------------------------------------------+
//	| X-Forward-User-Agent | Will be used as Diffbot's User-Agent header when making your request. |
//	| X-Forward-Referer    | Will be used as Diffbot's Referer header when making your request.    |
//	| X-Forward-Cookie     | Will be used as Diffbot's Cookie header when making your request.     |
//	| X-Forward-Accept-    | Will be used as Diffbot's Accept-Language header when making your     |
//	| Language             | request.                                                              |
//	+----------------------+-----------------------------------------------------------------------+
//
// Posting Content
//
// If your content is not publicly available (e.g., behind a firewall), you can POST markup directly to
// the Image API endpoint for analysis:
//
//  http://api.diffbot.com/v3/image?token=...&url=...
//
// Please note that the url argument is still required, and will be used to resolve any relative links
// contained in the markup.
//
// Provide the content to analyze as your POST body, and specify the Content-Type header as text/html.
//
// HTML Post Sample:
// curl -H "Content-Type: text/html" -d '<html><body><h2>Diffy the Robot</h2><div><img src="diffy-b.png"></div></body></html>' http://api.diffbot.com/v3/image?token=...&url=http%3A%2F%2Fwww.diffbot.com

type ImageResponse struct {
	Request *Request `json:"request"`
	Objects []*Image `json:"objects"`
}

func ParseImage(client *http.Client, token, url string, opt *Options) (*ImageResponse, error) {
	body, err := Diffbot(client, "image", token, url, opt)
	if err != nil {
		return nil, err
	}
	var result ImageResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (p *ImageResponse) String() string {
	d, _ := json.Marshal(p)
	return string(d)
}
