// Copyright 2014 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diffbot

import (
	"encoding/json"
	"net/http"
)

// See http://www.diffbot.com/dev/docs/product/
type Product struct {
	Type            string `json:"type"`
	PageUrl         string `json:"pageUrl"`
	ResolvedPageUrl string `json:"resolvedPageUrl,omitempty"`
	Title           string `json:"title"`
	HumanLanguage   string `json:"humanLanguage,omitempty"`
	DiffbotUri      string `json:"diffbotUri"`

	Text                string                 `json:"text,omitempty"`
	Images              []*productImageType    `json:"images,omitempty"`
	Discussion          Discussion             `json:"discussion"`
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

	// optional product fields
	Availability bool     `json:"availability,omitempty"`
	Colors       []string `json:"colors,omitempty"`
	Size         []string `json:"size,omitempty"`

	// optional fields
	Breadcrumb  []*breadcrumb          `json:"breadcrumb,omitempty"`
	Links       []string               `json:"links,omitempty"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
	QueryString string                 `json:"querystring,omitempty"`
}

// type of Product.PriceRange
type productPriceRange struct {
	MinPrice string `json:"minPrice,omitempty"`
	MaxPrice string `json:"maxPrice,omitempty"`
}

// type of Product.QuantityPrices
type productQuantityPrices struct {
	MinQuantity string `json:"minQuantity,omitempty"`
	Price       string `json:"price,omitempty"`
}

// type of Product.OfferPriceDetails, Product.RegularPriceDetails, Product.SaveAmountDetails
type productPriceDetails struct {
	Amount     float64 `json:"amount"` // number?
	Symbol     string  `json:"symbol"`
	Text       string  `json:"text"`
	Percentage bool    `json:"percentage,omitempty"`
}

// type of Product.Images[?]
type productImageType struct {
	URL           string `json:"url"`
	Title         string `json:"title"`
	NaturalHeight int    `json:"height"`
	NaturalWidth  int    `json:"width"`
	Primary       bool   `json:"primary,omitempty"`
	XPath         string `json:"xpath"`
	DiffbotUri    string `json:"diffbotUri"`
}

// The Product API automatically extracts complete data from any shopping or e-commerce product page.
// Retrieve full pricing information, product IDs (SKU, UPC, MPN), images, product specifications,
// brand and more.
//
// Request
//
// To use the Product API, perform a HTTP GET request on the following endpoint:
//
//  http://api.diffbot.com/v3/product
//
// Provide the following arguments:
//
//	+----------+-------------------------------------------------------------------------+
//	| ARGUMENT | DESCRIPTION                                                             |
//	+----------+-------------------------------------------------------------------------+
//	| token    | Developer token                                                         |
//	| url      | Web page URL of the product to process (URL encoded)                    |
//	+----------+-------------------------------------------------------------------------+
//	| Optional arguments                                                                 |
//	+----------+-------------------------------------------------------------------------+
//	| fields   | Used to specify optional fields to be returned by the Product API. See  |
//	|          | the Fields section below.                                               |
//	|discussion| Pass discussion=false to disable automatic extraction of product        |
//	|          | reviews. See below.                                                     |
//	| timeout  | Sets a value in milliseconds to wait for the retrieval/fetch of content |
//	|          | from the requested URL. The default timeout for the third-party         |
//	|          | response is 30 seconds (30000).                                         |
//	| callback | Use for jsonp requests. Needed for cross-domain ajax.                   |
//	+----------+-------------------------------------------------------------------------+
//
// The fields argument
// Use the fields argument to return optional fields in the JSON response. The default fields
// will always be returned. For nested arrays, use parentheses to retrieve specific fields, or
// * to return all sub-fields.
//
// For example, to return links and meta (in addition to the default fields), your &fields
// argument would be:
//
// &fields=links,meta
//
// Response
//
// The Product API returns data in JSON format.
//
// Each V3 response includes a request object (which returns request-specific metadata),
// and an objects array, which will include the extracted information for all objects on a
// submitted page.
//
// Objects in the Product API's objects array will include the following fields:
//
//	+-------------------+-------------------------------------------------------------------+
//	| FIELD             | DESCRIPTION                                                       |
//	+-------------------+-------------------------------------------------------------------+
//	| type              | Type of object (always product).                                  |
//	| pageUrl           | URL of submitted page / page from which the product is extracted. |
//	|resolvedPageUrl    | Returned if the pageUrl redirects to another URL.                 |
//	| title             | Title of the product.                                             |
//	| text              | Text description, if available, of the product.                   |
//	| brand             | Item's brand name.                                                |
//	| offerPrice        | Offer or actual/final price of the product.                       |
//	| regularPrice      | Regular or original price of the product, if available.           |
//	| shippingAmount    | Shipping price.                                                   |
//	| saveAmount        | Discount or amount saved off the regular price.                   |
//	| priceRange        | If the product is available in a range of prices, the minimum and |
//	|  |                | maximum values will be returned. The lowest price will also be    |
//	|  |                | returned as the offerPrice.                                       |
//	|  +- minPrice      | The minimum price for the offered item.                           |
//	|  +- maxPrice      | The maximum price for the offered item.                           |
//	| quantityPrices    | If the product is available with quantity-based discounts, all    |
//	|  |                | identifiable price points will be returned. The lowest price will |
//	|  |                | also be returned as the offerPrice.                               |
//	|  +-minQuantity    | The minimum quantity required to purchase for the associated      |
//	|  |                | price.                                                            |
//	|  +- price         | Price of the specific quantity level.                             |
//	| offerPriceDetails | offerPrice separated into its constituent parts: amount, symbol,  |
//	|                   | and full text.                                                    |
//	|regularPriceDetails| regularPrice separated into its constituent parts: amount, symbol,|
//	|                   | and full text.                                                    |
//	| saveAmountDetails	| saveAmount separated into its constituent parts: amount, symbol,  |
//	|                   | full text, and whether or not it is a percentage value.           |
//	| productId         | Diffbot-determined unique product ID. If upc, isbn, mpn or sku    |
//	|                   | are identified on the page, productId will select from these      |
//	|                   | values in the above order.                                        |
//	| upc               | Universal Product Code (UPC/EAN), if available.                   |
//	| sku               | Stock Keeping Unit -- store/vendor inventory number or identifier.|
//	| mpn               | Manufacturer's Product Number.                                    |
//	| isbn              | International Standard Book Number (ISBN), if available.          |
//	| specs             | If a specifications table or similar data is available on the     |
//	|                   | product page, individual specifications will be returned in the   |
//	|                   | specs object as name/value pairs. Names will be normalized to     |
//	|                   | lowercase with spaces replaced by underscores, e.g.               |
//	|                   | display_resolution.                                               |
//	| images            | Array of images, if present within the product.                   |
//	|  +- url           | Fully resolved link to image. If the image SRC is encoded as      |
//	|  |                | base64 data, the complete data URI will be returned.              |
//	|  +- title         | Description or caption of the image.                              |
//	|  +- naturalHeight | Raw image height, in pixels.                                      |
//	|  +- naturalWidth  | Raw image width, in pixels.                                       |
//	|  +- primary       | Returns true if image is identified as primary based on visual    |
//	|  |                | analysis.                                                         |
//	|  +- xpath         | XPath expression identifying the image node.                      |
//	|  +- diffbotUri    | Internal ID used for indexing.                                    |
//	| discussion        | Product reviews, as extracted by the Diffbot Discussion API. See  |
//	|                   | below.                                                            |
//	| prefixCode        | Country of origin as identified by UPC/ISBN.                      |
//	| productOrigin     | If available, two-character ISO country code where the product    |
//	|                   | was produced.                                                     |
//	| humanLanguage     | Returns the (spoken/human) language of the submitted page, using  |
//	|                   | two-letter ISO 639-1 nomenclature.                                |
//	| diffbotUri        | Unique object ID. The diffbotUri is generated from the values of  |
//	|                   | various Product fields and uniquely identifies the object. This   |
//	|                   | can be used for deduplication.                                    |
//	+---------------------------------------------------------------------------------------+
//	| Optional fields, available using fields= argument                                     |
//	+---------------------------------------------------------------------------------------+
//	| links             | Returns a top-level object (links) containing all hyperlinks      |
//	|                   | found on the page.                                                |
//	| meta              | Returns a top-level object (meta) containing the full contents of |
//	|                   | page meta tags, including sub-arrays for OpenGraph tags, Twitter  |
//	|                   | Card metadata, schema.org microdata, and -- if available -- oEmbed|
//	|                   | metadata.                                                         |
//	| querystring       | Returns any key/value pairs present in the URL querystring. Items |
//	|                   | without a discrete value will be returned as true.                |
//	| breadcrumb        | Returns a top-level array (breadcrumb) of URLs and link text from |
//	|                   | page breadcrumbs.                                                 |
//	+---------------------------------------------------------------------------------------+
//	| The following fields are in an early beta stage:                                      |
//	+---------------------------------------------------------------------------------------+
//	| availability      | Item's availability, either true or false.                        |
//	| colors            | Returns array of product color options.                           |
//	| size              | Size(s) available, if identified on the page.                     |
//	+---------------------------------------------------------------------------------------+
//
// Review Extraction
//
// By default the Product API will attempt to extract user reviews from product pages, using
// integrated functionality from the Diffbot Discussion API. (This behavior can be disabled
// using the argument discussion=false.)
//
// Review data will be returned in the discussion object (nested within the primary product
// object). The full syntax for discussion data is available in the Discussion API documentation.
//
// Example Response
//
//  {
//    "request": {
//      "pageUrl": "http://store.livrada.com/collections/all/products/before-i-go-to-sleep",
//      "api": "product",
//      "options": [],
//      "fields": "title,text,offerPrice,regularPrice,saveAmount,pageUrl,images",
//      "version": 3
//    },
//    {
//    "objects": [
//      {
//        "type": "product",
//        "title": "Before I Go To Sleep",
//        "text": "Memories define us. So what if you lost yours every time you went to sleep? Your name, your identity, your past, even the people you love -- all forgotten overnight. And the one person you trust may be telling you only half the story. Before I Go To Sleep is a disturbing psychological thriller in which an amnesiac desperately tries to uncover the truth about who she is and who she can trust.",
//        "offerPrice": "$7.99",
//        "regularPrice": "$9.99",
//        "saveAmount": "$2.00",
//        "pageUrl": "http://store.livrada.com/collections/all/products/before-i-go-to-sleep",
//        "images": [
//          {
//            "title": "Before I Go to Sleep cover",
//            "url": "http://cdn.shopify.com/s/files/1/0184/6296/products/BeforeIGoToSleep_large.png?946",
//            "xpath": "/HTML[@class='no-js']/BODY[@id='page-product']/DIV[@class='content-frame']/DIV[@class='content']/DIV[@class='content-shop']/DIV[@class='row']/DIV[@class='span5']/DIV[@class='product-thumbs']/UL/LI[@class='first-image']/A[@class='single_image']/IMG",
//            "diffbotUri": "image|1|768070723"
//          }
//        ]
//        "diffbotUri": "product|1|937176621"
//      }
//    ],
//  }
//
// Authentication
//
// You can supply Diffbot with basic authentication credentials or custom HTTP headers (see below)
// to access intranet pages or other sites that require a login.
//
// Basic Authentication
//
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
//  | X-Forward-Accept-    | Will be used as Diffbot's Accept-Language header when making your     |
//  | Language             | request.                                                              |
//	+----------------------+-----------------------------------------------------------------------+
//
// Posting Content
//
// If your content is not publicly available (e.g., behind a firewall), you can POST markup directly to the
// Product API endpoint for analysis:
//
//  http://api.diffbot.com/v3/product?token=...&url=...
//
// Please note that the url argument is still required, and will be used to resolve any relative links
// contained in the markup.
//
// Provide the content to analyze as your POST body, and specify the Content-Type header as text/html.
//
// HTML Post Sample:
//
// curl -H "Content-Type: text/html" -d '<html><head><title>Something to Buy</title></head><body><h2>A Pair of Jeans</h2><div>Price: $31.99</div></body></html>' http://api.diffbot.com/v3/product?token=...&url=http%3A%2F%2Fstore.diffbot.com

type ProductResponse struct {
	Request *Request   `json:"request"`
	Objects []*Product `json:"objects"`
}

func ParseProduct(client *http.Client, token, url string, opt *Options) (*ProductResponse, error) {
	body, err := Diffbot(client, "product", token, url, opt)
	if err != nil {
		return nil, err
	}
	var result ProductResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (p *ProductResponse) String() string {
	d, _ := json.Marshal(p)
	return string(d)
}
