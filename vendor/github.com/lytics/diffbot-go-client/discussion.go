package diffbot

import (
	"encoding/json"
	"net/http"
)

// See https://www.diffbot.com/dev/docs/discussion/#response
type Discussion struct {
	Type            string            `json:"type"`
	PageUrl         string            `json:"pageUrl"`
	ResolvedPageUrl string            `json:"resolvedPageUrl,omitempty"`
	Title           string            `json:"title"`
	NumPosts        int               `json:"numPosts"`
	Posts           []*discussionPost `json:"posts"`
	Tags            []*articleTag     `json:"tags"`
	Participants    int               `json:"participants"`
	NumPages        int               `json:"numPages"`
	NextPage        string            `json:"nextPage"`
	NextPages       []string          `json:"nextPages"`
	Provider        string            `json:"provider,omitempty"`
	HumanLanguage   string            `json:"humanLanguage,omitempty"`
	RssUrl          string            `json:"rssUrl,omitempty"`
	DiffbotUri      string            `json:"diffbotUri"`

	// optional fields
	Sentiment   float64                `json:"sentiment,omitempty"`
	Links       []string               `json:"links,omitempty"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
	QueryString string                 `json:"querystring,omitempty"`
	Breadcrumb  []*breadcrumb          `json:"breadcrumb,omitempty"`
}

// type of Discussion.Posts[?]
type discussionPost struct {
	Type          string        `json:"type"`
	Id            int           `json:"id"`
	ParentId      int           `json:"parentId,omitempty"`
	Text          string        `json:"text"`
	Html          string        `json:"html"`
	Tags          []*articleTag `json:"tags,omitempty"`
	Votes         int           `json:"votes,omitempty"`
	HumanLanguage string        `json:"humanLanguage"`
	Images        []*Image      `json:"image,omitempty"`
	Date          string        `json:"date"`
	Author        string        `json:"author"`
	AuthorUrl     string        `json:"authorUrl,omitempty"`
	PageUrl       string        `json:"pageUrl"`
	DiffbotUri    string        `json:"diffbotUri"`
}

// The Discussion API automatically structures and extracts entire threads or
// lists of reviews/comments from most discussion pages, forums, and similarly
// structured web pages.
//
// Request
//
// To use the Discussion Thread API, perform a HTTP GET request on the following
// endpoint:
//
//	http://api.diffbot.com/v3/discussion
//
// Provide the following arguments:
//
//	+----------+-----------------------------------------------------------------+
//	| ARGUMENT | DESCRIPTION                                                     |
//	+----------+-----------------------------------------------------------------+
//	| token    | Developer token                                                 |
//	| url      | Web page URL of the discussion to process (URL encoded)         |
//	+----------+-----------------------------------------------------------------+
//	| Optional arguments                                                         |
//	+----------+-----------------------------------------------------------------+
//	| fields   | Used to specify optional fields to be returned by the           |
//	|          | Discussion API. See the Fields section below.                   |
//	| timeout  | SSets a value in milliseconds to wait for the retrieval/fetch   |
//	|          | of content from the requested URL. The default timeout for the  |
//	|          | third-party response is 30 seconds (30000).                     |
//	| callback | Use for jsonp requests. Needed for cross-domain ajax.           |
//	| maxPages | Set the maximum number of pages in a thread to automatically    |
//	|          | concatenate in a single response. Default = 1 (no               |
//	|          | concatenation). Set maxPages=all to retrieve all pages of a     |
//	|          | thread regardless of length. Each individual page will count as |
//	|          | a separate API call.                                            |
//	+----------+-----------------------------------------------------------------+
//
// Response
//
// The Discussion API returns data in JSON format.
//
// Each V3 response includes a request object (which returns request-specific
// metadata), and an objects array, which will include the extracted information
// for all objects on a submitted page. The Discussion API returns all post data
// in a single object within the objects array.
//
// Within the Article and Product APIs (to extract comments or review data),
// discussion data will be returned within the nested discussion object.
//
// The Discussion API objects / discussion response will include the following fields:
//
//	+------------------+------------------------------------------------------------------------+
//	| FIELD            | DESCRIPTION                                                            |
//	+------------------+------------------------------------------------------------------------+
//	| type             | Type of object (always discussion).                                    |
//	| pageUrl          | URL of submitted page / page from which the discussion is extracted.   |
//	| resolvedPageUrl  | Returned if the pageUrl redirects to another URL.                      |
//	| title            | Title of the discussion.                                               |
//	| numPosts         | Number of individual posts in the thread.                              |
//	| posts            | Array of individual posts.                                             |
//	|  +- type         | Type of element (always post).                                         |
//	|  +- id           | ID of the individual post. The first post of a thread will have an ID  |
//	|  |               | of 0.                                                                  |
//	|  +- parentId     | ID of the parent, if the post is a reply or response.                  |
//	|  +- text         | Full text of the extracted post.                                       |
//	|  +- html         | Diffbot-normalized HTML of the extracted post. Please see the HTML     |
//	|  |               | Specification for a breakdown of elements and attributes returned.     |
//	|  +- tags         | If the post is long enough, an array of tags generated from its        |
//	|  |               | specific content.                                                      |
//	|  +-humanLanguage | Spoken/human language of the post, using two-letter ISO 639-1          |
//	|  |               | nomenclature.                                                          |
//	|  +- images       | If any images are detected within post content, they will be returned  |
//	|  |               | in a separate array. Individual array fields are the same as the       |
//	|  |               | Article API's images array.                                            |
//	|  +- date         | Date of post, normalized in most cases to RFC 1123 (HTTP/1.1).         |
//	|  +- author       | Name/username of the post author.                                      |
//	|  +- authorUrl    | URL of the author profile page, if available.                          |
//	|  +- pageUrl      | URL of the page on which the post was found.                           |
//	|  +- diffbotUri   | Internal ID used for indexing.                                         |
//	| tags             | Array of tags/entities as generated from analysis of all extracted     |
//	|                  | posts and cross-referenced with DBpedia and other data sources.        |
//	| participants     | Number of unique participants in the discussion thread or comments.    |
//	| numPages         | Number of pages in the thread concatenated to form the posts response. |
//	|                  | Use maxPages to define how many pages to concatenate.                  |
//	| nextPage         | If discussion spans multiple pages, nextPage will return the subsequent|
//	|                  | page URL.                                                              |
//	| nextPages        | Array of all page URLs concatenated in a multipage discussion.         |
//	| provider         | Discussion service provider (e.g., Disqus, Facebook), if known.        |
//	| humanLanguage    | Returns the (spoken/human) language of the submitted page, using       |
//	|                  | two-letter ISO 639-1 nomenclature.                                     |
//	| rssUrl           | URL of the discussion's RSS feed, if available.                        |
//	| diffbotUri       | Unique object ID. The diffbotUri is generated from the values of       |
//	|                  | various Discussion fields and uniquely identifies the object. This can |
//	|                  | be used for deduplication.                                             |
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
//	| breadcrumb       | Returns a top-level array (breadcrumb) of URLs and link text from page |
//	|                  | breadcrumbs.                                                           |
//	+------------------+------------------------------------------------------------------------+
//
// Example Response
// The following response shows the extracted contents from this Hacker News discussion thread:
//
//  {
//    "request": {
//      "pageUrl": "https://news.ycombinator.com/item?id=5608988",
//      "api": "discussion",
//      "version": 3
//    },
//    "objects": [
//      {
//        "title": "Show HN: Analysis of 2.5 Years of Frontpage Articles",
//        "numPosts": 7,
//        "diffbotUri": "discussion|3|-110040828",
//        "pageUrl": "https://news.ycombinator.com/item?id=5608988",
//        "posts": [
//          {
//            "id": 0,
//            "author": "johncoogan",
//            "text": "Huge fan of DiffBot and awesome projects like this. Really cool analysis, thanks for posting.\nWould be possible for you to post / send me the original data? I have been very interested in working on more longitudinal analysis using DiffBot data and this seems like a fun and interesting place to start. I'm happy to open-source / clearly attribute DiffBot's contribution to whatever I find / hack together, and would feel a lot more comfortable about integrating DiffBot into larger projects in the future.\nPlease email me (in my profile) if this is a possibility. Thanks!\n\n\n-----",
//            "diffbotUri": "post|3|-1215426659",
//            "pageUrl": "https://news.ycombinator.com/item?id=5608988",
//            "authorUrl": "https://news.ycombinator.com/user?id=johncoogan",
//            "html": "<p>Huge fan of DiffBot and awesome projects like this. Really cool analysis, thanks for posting.</p>\n<p>Would be possible for you to post / send me the original data? I have been very interested in working on more longitudinal analysis using DiffBot data and this seems like a fun and interesting place to start. I'm happy to open-source / clearly attribute DiffBot's contribution to whatever I find / hack together, and would feel a lot more comfortable about integrating DiffBot into larger projects in the future.</p>\n<p>Please email me (in my profile) if this is a possibility. Thanks!</p>\n<p>----- </p>",
//            "humanLanguage": "en",
//            "type": "post"
//          },
//          {
//            "id": 1,
//            "author": "tswaterman",
//            "parentId": 0,
//            "text": "Great idea! We'd be happy to share/help. If more people are interested, we'll figure out a good way to distribute the dataset generally. But in fact, you can extract the same data set, and add whatever other smart things you want along with it, using the Diffbot APIs. Everything we did to get this information is explained on our blog at\nhttp://blog.diffbot.com/diffbots-hackernews-trend-analyzer/ Sounds like you're already using the Diffbot service, but for anyone who's not, they can sign up for a free access token on our 'pricing' page at \nhttp://www.diffbot.com/ It's a few hundred thousand pages you'd need to analyze to get this, which doesn't quite fit under the free plan. You might not want to analyze as many years worth of stuff as we did for this demo, though.\nAll the pieces and services we used for this, including all the text extraction, topic detection, and crawling, are available to any user.\nHave fun with it, and keep us informed about whatever cool stuff you build with it, and of course tell us about any features or capabilities you wish Diffbot can provide.\n\n\n-----",
//            "diffbotUri": "post|3|-454130110",
//            "pageUrl": "https://news.ycombinator.com/item?id=5608988",
//            "authorUrl": "https://news.ycombinator.com/user?id=tswaterman",
//            "html": "<p>Great idea! We'd be happy to share/help. If more people are interested, we'll figure out a good way to distribute the dataset generally. But in fact, you can extract the same data set, and add whatever other smart things you want along with it, using the Diffbot APIs. Everything we did to get this information is explained on our blog at</p>\n<pre><code>    http://blog.diffbot.com/diffbots-hackernews-trend-analyzer/\n</code></pre>\n<p>Sounds like you're already using the Diffbot service, but for anyone who's not, they can sign up for a free access token on our 'pricing' page at <a href=\"http://www.diffbot.com/\">http://www.diffbot.com/</a> It's a few hundred thousand pages you'd need to analyze to get this, which doesn't quite fit under the free plan. You might not want to analyze as many years worth of stuff as we did for this demo, though.</p>\n<p>All the pieces and services we used for this, including all the text extraction, topic detection, and crawling, are available to any user.</p>\n<p>Have fun with it, and keep us informed about whatever cool stuff you build with it, and of course tell us about any features or capabilities you wish Diffbot can provide.</p>\n<p>----- </p>",
//            "humanLanguage": "en",
//            "type": "post"
//          },
//          {
//            "id": 2,
//            "author": "tliou",
//            "text": "Had to figure out how to use it ... but interesting once you do! Android vs IPhone on Hackernews frontpage shows spike in iphone on launch dates, but mediocre to no activity for android. is it because android is less interesting and not as innovative? or not as fun to talk/read about?\nhttp://diffbot.com/robotlab/hackernews/#type=tags&item=I...\n\n\n-----",
//            "diffbotUri": "post|3|-593417890",
//            "pageUrl": "https://news.ycombinator.com/item?id=5608988",
//            "authorUrl": "https://news.ycombinator.com/user?id=tliou",
//            "html": "<p>Had to figure out how to use it ... but interesting once you do! Android vs IPhone on Hackernews frontpage shows spike in iphone on launch dates, but mediocre to no activity for android. is it because android is less interesting and not as innovative? or not as fun to talk/read about?</p>\n<p><a href=\"http://diffbot.com/robotlab/hackernews/#type=tags&item=IPhone&item=Android%20(operating%20system)&item=\">http://diffbot.com/robotlab/hackernews/#type=tags&amp;item=I...</a></p>\n<p>----- </p>",
//            "humanLanguage": "en",
//            "type": "post"
//          },
//          {
//            "id": 3,
//            "author": "mayank",
//            "text": "Funny, I just built a HN article catcher that uses Diffbot to collect and classify submissions from the /new page [1]. I've been a Diffbot fan for quite a while now (although their entity recognition/tag classifier needs a bit of work as you can see from the categorization on my catcher page below).[1] http://lahiri.me/more\nI should add that their API is fantastic, and far better than using BeautifulSoup/NLTK for extracting textual content from webpages.\n\n\n-----",
//            "diffbotUri": "post|3|-1011030238",
//            "pageUrl": "https://news.ycombinator.com/item?id=5608988",
//            "authorUrl": "https://news.ycombinator.com/user?id=mayank",
//            "html": "<p>Funny, I just built a HN article catcher that uses Diffbot to collect and classify submissions from the /new page [1]. I've been a Diffbot fan for quite a while now (although their entity recognition/tag classifier needs a bit of work as you can see from the categorization on my catcher page below).</p>\n<p>[1] <a href=\"http://lahiri.me/more\">http://lahiri.me/more</a></p>\n<p>I should add that their API is fantastic, and far better than using BeautifulSoup/NLTK for extracting textual content from webpages.</p>\n<p>----- </p>",
//            "humanLanguage": "en",
//            "type": "post"
//          },
//          {
//            "id": 4,
//            "author": "tswaterman",
//            "parentId": 3,
//            "text": "Cool! How many articles, or what time period, did you use for this? It looks like you're using only a subset of the topic tags -- did you make a list of 'interesting stuff' to filter against?",
//            "diffbotUri": "post|3|-1575136385",
//            "pageUrl": "https://news.ycombinator.com/item?id=5608988",
//            "authorUrl": "https://news.ycombinator.com/user?id=tswaterman",
//            "html": "<p>Cool! How many articles, or what time period, did you use for this? It looks like you're using only a subset of the topic tags -- did you make a list of 'interesting stuff' to filter against?</p>",
//            "humanLanguage": "en",
//            "type": "post"
//          },
//          {
//            "id": 5,
//            "author": "mayank",
//            "parentId": 4,
//            "text": "It's been running for about a week I think, and I'm just taking the top 80 or so tags by article count. Glad you like it :)",
//            "diffbotUri": "post|3|-780525009",
//            "pageUrl": "https://news.ycombinator.com/item?id=5608988",
//            "authorUrl": "https://news.ycombinator.com/user?id=mayank",
//            "html": "<p>It's been running for about a week I think, and I'm just taking the top 80 or so tags by article count. Glad you like it :)</p>",
//            "humanLanguage": "en",
//            "type": "post"
//          },
//          {
//            "id": 6,
//            "author": "minimax",
//            "text": "Neat! Wish I could select by just domain name (i.e. just nytimes.com rather than dozen or so whatever.nytimes.com subdomains).",
//            "diffbotUri": "post|3|-458628829",
//            "pageUrl": "https://news.ycombinator.com/item?id=5608988",
//            "authorUrl": "https://news.ycombinator.com/user?id=minimax",
//            "html": "<p>Neat! Wish I could select by just domain name (i.e. just nytimes.com rather than dozen or so whatever.nytimes.com subdomains).</p>",
//            "humanLanguage": "en",
//            "type": "post"
//          }
//        ],
//        "humanLanguage": "en",
//        "confidence": 0.057376677520927157,
//        "numPages": 1,
//        "type": "discussion",
//        "participants": 5
//      }
//    ]
//  }
//
// Authentication and Custom Headers
//
// You can supply Diffbot with custom headers, or basic authentication credentials, in order to
// access intranet pages or other sites that require a login.
//
// Basic Authentication
// To access pages that require a login/password (using basic access authentication), include the
// username and password in your url parameter, e.g.: url=http%3A%2F%2FUSERNAME:PASSWORD@www.diffbot.com.
//
// Custom Headers
//
// You can supply the Discussion API with custom values for the user-agent, referer, or cookie values
// in the HTTP request. These will be used in place of the Diffbot default values.
//
// To provide custom headers, pass in the following values in your own headers when calling the Diffbot API:
//	+----------------------+-----------------------------------------------------------------------+
//	| HEADER               | DESCRIPTION                                                           |
//	+----------------------+-----------------------------------------------------------------------+
//	| X-Forward-User-Agent | Will be used as Diffbot's User-Agent header when making your request. |
//	| X-Forward-Referer    | Will be used as Diffbot's Referer header when making your request.    |
//	| X-Forward-Cookie     | Will be used as Diffbot's Cookie header when making your request.     |
//	+----------------------+-----------------------------------------------------------------------+
//
// Posting Content
//
// If your content is not publicly available (e.g., behind a firewall), you can POST markup directly
// to the Discussion API endpoint for analysis:
//
//  http://api.diffbot.com/v3/discussion?token=...&url=...
//
// Please note that the url argument is still required, and will be used to resolve any relative links
// contained in the markup.
//
// Provide the content to analyze as your POST body, and specify the Content-Type header as text/html.

func ParseDiscussion(client *http.Client, token, url string, opt *Options) (*Discussion, error) {
	body, err := Diffbot(client, "discussion", token, url, opt)
	if err != nil {
		return nil, err
	}
	var result Discussion
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (p *Discussion) String() string {
	d, _ := json.Marshal(p)
	return string(d)
}
