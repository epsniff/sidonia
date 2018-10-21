// Copyright 2014 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diffbot

import (
	"encoding/json"
	"net/http"
)

// See http://diffbot.com/dev/docs/analyze/
type Classification struct {
	// common properties
	Type            string              `json:"type"`
	Title           string              `json:"title"`
	PageUrl         string              `json:"pageUrl"`
	ResolvedPageUrl string              `json:"resolvedPageUrl,omitempty"`
	DiffbotUri      string              `json:"diffbotUri"`
	Text            string              `json:"text,omitempty"`
	HumanLanguage   string              `json:"humanLanguage,omitempty"`
	Images          []*articleImageType `json:"images,omitempty"`
	Discussion      Discussion          `json:"discussion"`
	Tags            []*articleTag       `json:"tags"`
	NumPages        int                 `json:"numPages"`
	NextPages       []string            `json:"nextPages"`
	Url             string              `json:"url"`
	Html            string              `json:"html"`
	Author          string              `json:"author"`
	Date            string              `json:"date"`
	NaturalHeight   int                 `json:"naturalHeight,omitempty"`
	NaturalWidth    int                 `json:"naturalWidth,omitempty"`

	// optional fields
	Breadcrumb  []*breadcrumb          `json:"breadcrumb,omitempty"`
	Links       []string               `json:"links,omitempty"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
	QueryString string                 `json:"querystring,omitempty"`

	// article properties
	EstimatedDate    string              `json:"estimatedDate,omitempty"`
	AuthorUrl        string              `json:"authorUrl,omitempty"`
	SiteName         string              `json:"siteName"`
	PublisherRegion  string              `json:"publisherRegion,omitempty"`
	PublisherCountry string              `json:"publisherCountry,omitempty"`
	Videos           []*articleVideoType `json:"videos"`

	// discussion properties
	NumPosts     int              `json:"numPosts"`
	Posts        []discussionPost `json:"posts"`
	Participants int              `json:"participants"`
	NextPage     string           `json:"nextPage"`
	Provider     string           `json:"provider,omitempty"`
	RssUrl       string           `json:"rssUrl,omitempty"`
	Sentiment    float64          `json:"sentiment,omitempty"`

	// image properties
	AnchorUrl     string                   `json:"anchorUrl,omitempty"`
	XPath         string                   `json:"xpath,omitempty"`
	DisplayHeight int                      `json:"displayHeight,omitempty"`
	DisplayWidth  int                      `json:"displayWidth,omitempty"`
	Mentions      []string                 `json:"mentions,omitempty"`
	Ocr           string                   `json:"ocr,omitempty"`
	Faces         []map[string]interface{} `json:"faces,omitempty"`

	// product properties
	Brand               string                 `json:"brand,omitempty"`
	OfferPrice          string                 `json:"offerPrice"`
	RegularPrice        string                 `json:"regularPrice,omitempty"`
	ShippingAmount      string                 `json:"shippingAmount"`
	SaveAmount          string                 `json:"saveAmount"`
	PriceRange          *productPriceRange     `json:"priceRange,omitempty"`
	QuantityPrices      *productQuantityPrices `json:"quantityPrices,omitempty"`
	OfferPriceDetails   *productPriceDetails   `json:"offerPriceDetails,omitempty"`
	RegularPriceDetails *productPriceDetails   `json:"regularPriceDetails,omitempty"`
	SaveAmountDetails   *productPriceDetails   `json:"saveAmountDetails,omitempty"`
	ProductId           string                 `json:"productId"`
	UPC                 string                 `json:"upc,omitempty"`
	SKU                 string                 `json:"sku,omitempty"`
	MPN                 string                 `json:"mpn,omitempty"`
	ISBN                string                 `json:"isbn,omitempty"`
	Specs               map[string]interface{} `json:"specs,omiyempty"`
	PrefixCode          string                 `json:"prefixCode"`
	ProductOrigin       string                 `josn:"productOrigin,omitempty"`
	Availability        bool                   `json:"availability,omitempty"`
	Colors              []string               `json:"colors,omitempty"`
	Size                []string               `json:"size,omitempty"`

	// video properties
	EmbedUrl  string `json:"embedUrl,omitempty"`
	Duration  int    `json:"duration"`
	ViewCount int    `json:"viewCount,omitempty"`
	Mime      string `json:"mime`
}

type breadcrumb struct {
	Link string `json:"link"`
	Name string `json:"name"`
}

// The Diffbot Analyze API visually analyzes a web page, identifies its "page-type," and determines which
// Diffbot extraction API (if any) is appropriate. Pages that match a supported extraction API -- articles,
// discussions, images, products or videos -- will be automatically extracted and returned in the Analyze API
// response.
//
// Pages not currently supported by an extraction API will return "other."
//
// Request
// To use the Analyze API, perform a HTTP GET request on the following endpoint:
//
//  http://api.diffbot.com/v3/analyze?token=...&url=...
//
// Provide the following arguments:
//
//	+----------+----------------------------------------------------------------------------------------------+
//	| ARGUMENT | DESCRIPTION                                                                                  |
//	+----------+----------------------------------------------------------------------------------------------+
//	| token    | Developer token                                                                              |
//	| url      | Web page URL of the analyze to process (URL encoded)                                         |
//	+----------+----------------------------------------------------------------------------------------------+
//	| Optional arguments                                                                                      |
//	+----------+----------------------------------------------------------------------------------------------+
//	| mode     | By default the Analyze API will fully extract all pages that match an existing Automatic API |
//	|          | -- articles, products or image pages. Set mode to a specific page-type (e.g., mode=article)  |
//	|          | to extract content only from that specific page-type. All other pages will simply return the |
//	|          | default Analyze fields.                                                                      |
//	| fallback | Force any non-extracted pages (those with a type of "other") through a specific API. For     |
//	|          | example, to route all "other" pages through the Article API, pass &fallback=article. Pages   |
//	|          | that utilize this functionality will return a fallbackType field at the top-level of the     |
//	|          | response, indicating the fallback API used.                                                  |
//	| fields   | Specify optional fields to be returned from any fully-extracted pages, e.g.:                 |
//	|          | &fields=querystring,links. See available fields within each API's individual documentation   |
//	|          | pages.                                                                                       |
//	|discussion| Pass discussion=false to disable automatic extraction of comments or reviews from pages      |
//	|          | identified as articles or products. This will not affect pages identified as discussions.    |
//	| timeout  | Sets a value in milliseconds to wait for the retrieval/fetch of content from the requested   |
//	|          | URL. The default timeout for the third-party response is 30 seconds (30000).                 |
//	| callback | Use for jsonp requests. Needed for cross-domain ajax.                                        |
//	+----------+----------------------------------------------------------------------------------------------+
//
// Response
//
// The Analyze API returns data in JSON format.
//
// Each response includes a request object (which returns request-specific metadata), and an objects array,
// which will include the extracted information for all objects on a submitted page.
//
// If the Analyze API identifies the submitted page as an article, discussion thread, product or image, the
// associated object(s) from the page will be returned automatically in the objects array.
//
// The default fields returned:
//
//	+----------------+------------------------------------------------------------------+
//	| FIELD          | DESCRIPTION                                                      |
//	+----------------+------------------------------------------------------------------+
//	| title          | Title of the page.                                               |
//	| type           | Page-type of the submitted URL, either article, image, product   |
//	|                | or other.                                                        |
//	| human_language | Returns the (spoken/human) language of the submitted URL,        |
//	|                | using two-letter ISO 639-1 nomenclature.                         |
//	|                | Returned by default.                                             |
//	+----------------+------------------------------------------------------------------+
//	| Optional fields, available using fields= argument                                 |
//	+----------------+------------------------------------------------------------------+
//	| links          | Returns a top-level object (links) containing all hyperlinks     |
//	|                | found on the page.                                               |
//	| meta           | Returns a top-level object (meta) containing the full contents   |
//	|                | of page meta tags, including sub-arrays for OpenGraph tags,      |
//	|                | Twitter Card metadata, schema.org microdata, and -- if available |
//	|                | -- oEmbed metadata.                                              |
//	| querystring    | Returns any key/value pairs present in the URL querystring.      |
//	|                | Items without a discrete value will be returned as true.         |
//	| breadcrumb     | Returns a top-level array (breadcrumb) of URLs and link text     |
//	|                | from page breadcrumbs.                                           |
//	+----------------+------------------------------------------------------------------+
//
// Example Response
//
// Because the below classified page is an article, its full contents are extracted using the Article API:
//
//  {
//    "request": {
//      "pageUrl": "http://tcrn.ch/Jw7ZKw",
//      "resolvedPageUrl": "http://techcrunch.com/2012/05/31/diffbot-raises-2-million-seed-round-for-web-content-extraction-technology/",
//      "api": "analyze",
//      "options": [],
//      "fields": "",
//      "version": 3
//    },
//    "objects": [
//      {
//        "type": "article",
//        "resolvedPageUrl": "http://techcrunch.com/2012/05/31/diffbot-raises-2-million-seed-round-for-web-content-extraction-technology/",
//        "pageUrl": "http://tcrn.ch/Jw7ZKw",
//        "human_language": "en",
//        "text": "Diffbot , the super-geeky/awesome visual learning robot technology which aims to \"see\" the web the way that people do, is today announcing a new infusion of capital. The company has closed $2 million in funding from a number of technology veterans, including EarthLink founder Sky Dayton ; Andy Bechtolsheim , co-founder of Sun Microsystems; Joi Ito , Director of MIT Media Lab; Brad Garlinghouse , CEO of YouSendIt ( and formerly of TechCrunch parent company AOL ), Maynard Webb , Chairman of the Board at LiveOps, formerly eBay COO; Elad Gil , VP of Corporate Strategy at Twitter; Jonathan Heiliger , former VP of Technical Operations at Facebook; Redbeacon co-founder Aaron Lee ; and founder of VitalSigns Montgomery Kersten .\nMatrix Partners also participated in the round. Of the new investors, Sky Dayton will be the first to join Diffbot's board and will be taking an active role in the company, including plans to go hands-on with various Diffbot projects.\nLast August, the company publicly debuted its first APIs , which allow developers to build apps that can automatically extract meaning from web pages. For example, the Front Page API is able to analyze site homepages, and understands the difference between article text, headlines, bylines, ads, etc. The Article API can then extract clean article text, images and videos. Another example of Diffbot in action is the \"follow API,\" which can track the changes made to a website.\nToday, Diffbot has categorized the web into about 20 different page types, including homepages and article pages, which are the first two types it can now identity. Going forward, Diffbot plans train its bots to recognize all the other types of pages, including product pages, social networking profiles, recipe pages, review pages, and more.\nIts APIs have been put to use by AOL (again: disclosure, TC parent) in its news magazine AOL Editions , as well as by companies like Nuance , SocMetrics , and others. Diffbot says it's now processing 100 million API calls per month on behalf of its customers. Thousands of developers are using the APIs, the company notes, but paying customers are only in the \"tens.\" Correction: we're now told they have \"a lot more!\"\nDiffbot founder and CEO Michael Tung (aka \"Diffbot Mike\") says the new funding will be put towards new hires and expanding its resources. "More than that, we're receiving a huge vote of confidence from veterans who have built massive companies and understand the fine points of building for scale, maintaining uptime and delivering the absolute highest standards of service."\nTung is a patent attorney and Stanford PhD student who left the doctoral program to pursue Diffbot, thanks to seed funding from Stanford's incubator, StartX . Diffbot was StartX's first investment. With today's funding, Diffbot total raise is $2 million and change.",
//        "title": "Diffbot Raises $2 Million Angel Round For Web Content Extraction Technology",
//        "images": [
//          {
//          "primary": "true",
//          "url": "http://tctechcrunch2011.files.wordpress.com/2012/05/diffbot_9.png?w=300"
//          }
//        ],
//        "date": "Thu, 31 May 2012 07:00:00 GMT"
//      }
//  }
//
// Authentication
//
// You can supply Diffbot with basic authentication credentials or custom HTTP headers (see below) to access
// intranet pages or other sites that require a login.
//
// Basic Authentication
// To access pages that require a login/password (using basic access authentication), include the username and
// password in your url parameter, e.g.: url=http%3A%2F%2FUSERNAME:PASSWORD@www.diffbot.com.
//
// Custom HTTP Headers
//
// You can supply Diffbot APIs with custom values for the user-agent, referer, cookie, or accept-language values
// in the HTTP request. These will be used in place of the Diffbot default values.
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

type ClassificationResponse struct {
	// default fields
	HumanLanguage string `json:"humanLanguage"`
	Title         string `json:"title"`
	Type          string `json:"type"`

	Request *Request          `json:"request"`
	Objects []*Classification `json:"objects"`
}

func ParseClassification(client *http.Client, token, url string, opt *Options) (*ClassificationResponse, error) {
	body, err := Diffbot(client, "analyze", token, url, opt)
	if err != nil {
		return nil, err
	}
	var result ClassificationResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (p *ClassificationResponse) String() string {
	d, _ := json.Marshal(p)
	return string(d)
}
