// Copyright 2014 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diffbot

import (
	"encoding/json"
	"net/http"
)

// See http://diffbot.com/dev/docs/article/
type Article struct {
	Type             string              `json:"type"`
	Title            string              `json:"title"`
	Text             string              `json:"text"`
	Html             string              `json:"html"`
	Date             string              `json:"date"`
	EstimatedDate    string              `json:"estimatedDate,omitempty"`
	Author           string              `json:"author"`
	AuthorUrl        string              `json:"authorUrl,omitempty"`
	Discussion       *Discussion         `json:"discussion"`
	HumanLanguage    string              `json:"humanLanguage,omitempty"`
	NumPages         int                 `json:"numPages"`
	NextPages        []string            `json:"nextPages"`
	SiteName         string              `json:"siteName"`
	PublisherRegion  string              `json:"publisherRegion,omitempty"`
	PublisherCountry string              `json:"publisherCountry,omitempty"`
	PageUrl          string              `json:"pageUrl"`
	ResolvedUrl      string              `json:"resolvedPageUrl,omitempty"`
	Tags             []*articleTag       `json:"tags,omitempty"`
	Images           []*articleImageType `json:"images"`
	Videos           []*articleVideoType `json:"videos"`
	DiffbotUri       string              `json:"diffbotUri"`

	// optional fields
	Breadcrumb  []*breadcrumb          `json:"breadcrumb,omitempty"`
	Links       []string               `json:"links,omitempty"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
	QueryString string                 `json:"querystring,omitempty"`
}

// type of Article.Tag[?]
type articleTag struct {
	Label    string   `json:"label"`
	Count    int      `json:"count"`
	Score    float64  `json:"score"`
	RdfTypes []string `json:"rdfTypes,omitempty"`
	Type     string   `json:"type,omitempty"`
	Uri      string   `json:"uri"`
}

// type of Article.Images[?]
type articleImageType struct {
	Url           string `json:"url"`
	Title         string `json:"title"`
	Height        int    `json:"height"`
	Width         int    `json:"width"`
	NaturalHeight int    `json:"naturalHeight"`
	NaturalWidth  int    `json:"naturalWidth"`
	Primary       bool   `json:"primary,omitempty"`
	DiffbotUri    string `json:"diffbotUri"`
}

// type of Article.Videos[?]
type articleVideoType struct {
	Url           string `json:"url"`
	NaturalHeight int    `json:"naturalHeight,omitempty"`
	NaturalWidth  int    `json:"naturalWidth,omitempty"`
	Primary       bool   `json:"primary,omitempty"`
	DiffbotUri    string `json:"diffbotUri"`
}

// The Article API is used to extract clean article text and other data from news
// articles, blog posts and other text-heavy pages. Retrieve the full-text, cleaned
// and normalized HTML, related images and videos, author, date, tags—automatically,
// from any article on any site.
//
// Request
//
// To use the Article API, perform a HTTP GET request on the following endpoint:
//
//  http://api.diffbot.com/v3/article
//
// Provide the following arguments:
//
//	+----------+-----------------------------------------------------------------+
//	| ARGUMENT | DESCRIPTION                                                     |
//	+----------+-----------------------------------------------------------------+
//	| token    | Developer token                                                 |
//	| url      | Web page URL of the article to process (URL encoded)            |
//	+----------+-----------------------------------------------------------------+
//	| Optional arguments                                                         |
//	+----------+-----------------------------------------------------------------+
//	| fields   | Used to specify optional fields to be returned by the Article   |
//	|          | API. See the Fields section below.                              |
//	| paging   | Pass paging=false to disable automatic concatenation of         |
//	|          | multiple-page articles. (By default, Diffbot will concatenate   |
//	|          | up to 20 pages of a single article.)                            |
//	| maxTags  | Set the maximum number of automatically-generated tags to       |
//	|          | return. By default a maximum of five tags will be returned.     |
//	|discussion| Pass discussion=false to disable automatic extraction of        |
//	|          | article comments.                                               |
//	| timeout  | Sets a value in milliseconds to wait for the retrieval/fetch of |
//	|          | content from the requested URL. The default timeout for the     |
//	|          | third-party response is 30 seconds (30000).                     |
//	| callback | Use for jsonp requests. Needed for cross-domain ajax.           |
//	+----------+-----------------------------------------------------------------+
//
// The fields argument
//
// Use the fields argument to return optional fields in the JSON response. The
// default fields will always be returned. For nested arrays, use parentheses to
// retrieve specific fields, or * to return all sub-fields.
//
// For example, to return links and meta (in addition to the default fields), your
// &fields argument would be:
//
//  &fields=links,meta
//
// Response
//
// The Article API returns data in JSON format.
//
// Each V3 response includes a request object (which returns request-specific metadata),
// and an objects array, which will include the extracted information for all objects
// on a submitted page. At the moment, only a single object will be returned for
// Article API requests.
//
// Objects in the Article API's objects array will include the following fields:
//
//	+------------------+------------------------------------------------------------------------+
//	| FIELD            | DESCRIPTION                                                            |
//	+------------------+------------------------------------------------------------------------+
//	| type             | Type of object (always article).                                       |
//	| title            | Title of the article.                                                  |
//	| text             | Full text of the article.                                              |
//	| html             | Diffbot-normalized HTML of the extracted article. Please see the HTML  |
//	|                  | Specification for a breakdown of elements and attributes returned.     |
//	| date             | Date of extracted article, normalized in most cases to RFC 1123        |
//	|                  | (HTTP/1.1).                                                            |
//	| estimatedDate    | If an article's date is ambiguous, Diffbot will attempt to estimate a  |
//	|                  | more specific timestamp using various factors. This will not be        |
//	|                  | generated for articles older than two days, or articles without an     |
//	|                  | identified date.                                                       |
//	| author           | Article author.                                                        |
//	| authorUrl        | URL of the author profile page, if available.                          |
//	| discussion       | Article comments, as extracted by the Diffbot Discussion API.          |
//	| humanLanguage    | Returns the (spoken/human) language of the submitted page, using       |
//	|                  | two-letter ISO 639-1 nomenclature.                                     |
//	| numPages         | Number of pages automatically concatenated to form the text or html    |
//	|                  | esponse. By default, Diffbot will automatically concatenate up to 20   |
//	|                  | pages of an article.                                                   |
//	| nextPages        | Array of all page URLs concatenated in a multipage article.            |
//	| siteName         | The plain-text name of the site (e.g. The New York Times or Diffbot).  |
//	|                  | If no site name is automatically determined, the root domain           |
//	|                  | (diffbot.com) will be returned.                                        |
//	| publisherRegion  | If known, the region of the article publication.                       |
//	| publisherCountry | If known, the country of the article publication.                      |
//	| pageUrl          | URL of submitted page / page from which the article is extracted.      |
//	| resolvedPageUrl  | Returned if the pageUrl redirects to another URL.                      |
//	| tags             | Array of tags/entities, generated from analysis of the extracted text  |
//	|  |               | and cross-referenced with DBpedia and other data sources.              |
//	|  +- label        | Name of the entity or tag.                                             |
//	|  +- count        | Number of appearances the entity makes within the text content.        |
//	|  +- score        | Rating of the entity's relevance to the overall text content (range of |
//	|  |               | 0 to 1) based on various factors.                                      |
//	|  +- rdfTypes     | If the entity can be represented by multiple resources, all of the     |
//	|  |               | possible URIs will be returned.                                        |
//	|  +- type         | Simplified type, if determined (e.g. organization or person).          |
//	|  +- url          | Link to the primary entity at DBpedia or other data source.            |
//	| images           | Array of images, if present within the article body.                   |
//	|  +- url          | Fully resolved link to image. If the image SRC is encoded as base64    |
//	|  |               | data, the complete data URI will be returned.                          |
//	|  +- title        | Description or caption of the image.                                   |
//	|  +- height       | Height of image as (re-)sized via browser/CSS.                         |
//	|  +- width        | Width of image as (re-)sized via browser/CSS.                          |
//	|  +-naturalHeight | Raw image height, in pixels.                                           |
//	|  +-naturalWidth  | Raw image width, in pixels.                                            |
//	|  +- primary      | Returns true if image is identified as primary based on visual         |
//	|  |               | analysis.                                                              |
//	|  +- diffbotUri   | Internal ID used for indexing.                                         |
//	| videos           | Array of videos, if present within the article body.                   |
//	|  +- url          | Fully resolved link to source video content.                           |
//	|  +-naturalHeight | Source video height, in pixels, if available.                          |
//	|  +-naturalWidth  | Source video width, in pixels, if available.                           |
//	|  +- primary      | Returns true if video is identified as primary based on visual         |
//	|  |               | analysis.                                                              |
//	|  +- diffbotUri   | Internal ID used for indexing.                                         |
//	| breadcrumb       | Returns a top-level array (breadcrumb) of URLs and link text from page |
//	|                  | breadcrumbs.                                                           |
//	| diffbotUri       | Unique object ID. The diffbotUri is generated from the values of       |
//	|                  | various Article fields and uniquely identifies the object. This can be |
//	|                  | used for deduplication.                                                |
//	+-------------------------------------------------------------------------------------------+
//	| Optional fields, available using fields= argument                                         |
//	+-------------------------------------------------------------------------------------------+
//	| sentiment        | Returns the sentiment score of the analyzed article text, a value      |
//	|                  | ranging from -1.0 (very negative) to 1.0 (very positive).              |
//	| links            | Returns a top-level object (links) containing all hyperlinks found on  |
//	|                  | the page.                                                              |
//	| meta             | Returns a top-level object (meta) containing the full contents of page |
//	|                  | meta tags, including sub-arrays for OpenGraph tags, Twitter Card       |
//	|                  | metadata, schema.org microdata, and -- if available -- oEmbed metadata.|
//	| querystring      | Returns any key/value pairs present in the URL querystring. Items      |
//	|                  | without a discrete value will be returned as true.                     |
//	+------------------+------------------------------------------------------------------------+
//
// Comment Extraction
//
// By default the Article API will attempt to extract comments from article pages, using
// integrated functionality from the Diffbot Discussion API. (This behavior can be disabled
// using the argument discussion=false.)
//
// Comment data will be returned in the discussion object (nested within the primary article
// object). The full syntax for discussion data is available in the Discussion API documentation.
//
// Example Response
//
// The following request --
//  http://api.diffbot.com/v3/article?token=...&url=http%3A%2F%2Fblog.diffbot.com%2Fdiffbots-new-product-api-teaches-robots-to-shop-online
//
// -- will result in this API response:
//
//  {
//    "request": {
//      "pageUrl": "http://blog.diffbot.com/diffbots-new-product-api-teaches-robots-to-shop-online",
//      "api": "article",
//      "version": 3
//    },
//    "objects": [
//      {
//        "date": "Wed, 31 Jul 2013 00:00:00 GMT",
//        "author": "John Davi",
//        "estimatedDate": "Wed, 31 Jul 2013 00:00:00 GMT",
//        "publisherRegion": "North America",
//        "diffbotUri": "article|3|-820542508",
//        "siteName": "Diffbot",
//        "videos": [
//          {
//            "diffbotUri": "video|3|-761237582",
//            "url": "http://www.youtube.com/embed/lfcri5ungRo?feature=oembed",
//            "primary": true
//          }
//        ],
//        "type": "article",
//        "title": "Diffbot's New Product API Teaches Robots to Shop Online",
//        "tags": [
//          {
//            "score": 0.48,
//            "count": 1,
//            "label": "Online and offline",
//            "uri": "http://dbpedia.org/resource/Online_and_offline"
//          },
//          {
//            "score": 0.45,
//            "count": 1,
//            "label": "Software release life cycle",
//            "uri": "http://dbpedia.org/resource/Software_release_life_cycle"
//          },
//          {
//            "score": 0.51,
//            "count": 2,
//            "label": "Structured content",
//            "uri": "http://dbpedia.org/resource/Structured_content"
//          },
//          {
//            "score": 0.5,
//            "count": 3,
//            "label": "Data",
//            "uri": "http://dbpedia.org/resource/Data"
//          },
//          {
//            "score": 0.78,
//            "count": 5,
//           "label": "Application programming interface",
//            "uri": "http://dbpedia.org/resource/Application_programming_interface"
//          }
//        ],
//        "publisherCountry": "Diffbot HQ",
//        "humanLanguage": "en",
//        "authorUrl": "http://blog.diffbot.com/author/johndavi/",
//        "pageUrl": "http://blog.diffbot.com/diffbots-new-product-api-teaches-robots-to-shop-online",
//        "html": "<p>Diffbot&rsquo;s human wranglers are proud today to announce the release of our newest product: an API for&hellip; products!</p>\n<p>The <a href=\"http://www.diffbot.com/products/automatic/product\">Product API</a> can be used for extracting clean, structured data from any e-commerce product page. It automatically makes available all the product data you&rsquo;d expect: price, discount/savings amount, shipping cost, product description, any relevant product images, SKU and/or other product IDs.</p>\n<p>Even cooler: pair the Product API with <a href=\"http://www.diffbot.com/products/crawlbot\">Crawlbot</a>, our intelligent site-spidering tool, and let Diffbot determine which pages are products, then automatically structure the entire catalog. Here&rsquo;s a quick demonstration of Crawlbot at work:</p>\n<figure><iframe frameborder=\"0\" src=\"http://www.youtube.com/embed/lfcri5ungRo?feature=oembed\"></iframe></figure>\n<p>We&rsquo;ve developed the Product API over the course of two years, building upon our core vision technology that&rsquo;s extracted structured data from billions of web pages, and training our machine learning systems using data from tens of thousands of unique shopping sites. We can&rsquo;t wait for you to try it out.</p>\n<p>What are you waiting for? Check out the <a href=\"http://www.diffbot.com/products/automatic/product\">Product API documentation</a> and dive on in! If you need a token, check out our <a href=\"http://www.diffbot.com/pricing\">pricing and plans</a> (including our Free plan).</p>\n<p>Questions? Hit us up at <a href=\"mailto:support@diffbot.com\">support@diffbot.com</a>.</p>",
//        "text": "Diffbot's human wranglers are proud today to announce the release of our newest product: an API for\u2026 products!\nThe Product API can be used for extracting clean, structured data from any e-commerce product page. It automatically makes available all the product data you'd expect: price, discount/savings amount, shipping cost, product description, any relevant product images, SKU and/or other product IDs.\nEven cooler: pair the Product API with Crawlbot, our intelligent site-spidering tool, and let Diffbot determine which pages are products, then automatically structure the entire catalog. Here's a quick demonstration of Crawlbot at work:\nWe've developed the Product API over the course of two years, building upon our core vision technology that's extracted structured data from billions of web pages, and training our machine learning systems using data from tens of thousands of unique shopping sites. We can't wait for you to try it out.\nWhat are you waiting for? Check out the Product API documentation and dive on in! If you need a token, check out our pricing and plans (including our Free plan).\nQuestions? Hit us up at support@diffbot.com.",
//        "resolvedPageUrl": "http://blog.diffbot.com/diffbots-new-product-api-teaches-robots-to-shop-online/"
//      }
//    ]
//  }
//
// Authentication
//
// You can supply Diffbot with basic authentication credentials or custom HTTP headers
// (see below) to access intranet pages or other sites that require a login.
//
//	Basic Authentication
//
//	To access pages that require a login/password (using basic access authentication),
//	include the username and password in your url parameter,
//  e.g.: url=http%3A%2F%2FUSERNAME:PASSWORD@www.diffbot.com.
//
//	Custom HTTP Headers
//
//	You can supply Diffbot APIs with custom values for the user-agent, referer, cookie, or
//  accept-language values in the HTTP request. These will be used in place of the Diffbot
//  default values
//
// To provide custom headers, pass in the following values in your own headers when calling
// the Diffbot API:
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
// If your content is not publicly available (e.g., behind a firewall), you can POST markup
// or plain text directly to the Article API endpoint for analysis:
//
//	http://api.diffbot.com/v3/article?token=...&url=...
//
// Please note that if you submit HTML, the url argument is still required, and will be used
// to resolve any relative links contained in the markup.
//
// Provide the content to analyze as your POST body, and specify the Content-Type header as
// text/html (for full markup) or text/plain (for text-only).
//
// HTML Post Sample:
//	curl
//    -H "Content-Type: text/html"
//    -d '<html><body><p>Now is the time for all good robots to come to the aid of their-- oh never mind, run!</p></body></html>'
//    http://api.diffbot.com/v3/article?token=...&url=http%3A%2F%2Fblog.diffbot.com
//
// Plaintext Post Sample:
// curl -H "Content-Type: text/plain" -d 'Now is the time for all good robots to come to the aid of their-- oh never mind, run!' http://api.diffbot.com/v3/article?token=...&fields=tags,text
//

type ArticleResponse struct {
	Request *Request   `json:"request"`
	Objects []*Article `json:"objects"`
}

func ParseArticle(client *http.Client, token, url string, opt *Options) (*ArticleResponse, error) {
	body, err := Diffbot(client, "article", token, url, opt)
	if err != nil {
		return nil, err
	}
	var result ArticleResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (p *ArticleResponse) String() string {
	d, _ := json.Marshal(p)
	return string(d)
}
