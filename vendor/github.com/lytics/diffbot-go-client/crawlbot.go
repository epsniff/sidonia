package diffbot

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// See https://www.diffbot.com/dev/docs/crawl
type Crawl struct {
	Name                          string         `json:"name"`
	Type                          string         `json:"type"`
	JobCreationTimeUTC            int            `json:"jobCreationTimeUTC"`
	JobCompletionTimeUTC          int            `json:"jobCompletionTimeUTC"`
	JobStatus                     *jobStatusType `json:"jobStatus"`
	SentJobDoneNotification       int            `json:"sentJobDoneNotification"`
	ObjectsFound                  int            `json:"objectsFound"`
	UrlsHarvested                 int            `json:"urlsHarvested"`
	PageCrawlAttempts             int            `json:"pageCrawlAttempts"`
	PageCrawlSuccesses            int            `json:"pageCrawlSuccesses"`
	PageCrawlSuccessesThisRound   int            `json:"pageCrawlSuccessesThisRound"`
	PageProcessAttempts           int            `json:"pageProcessAttempts"`
	PageProcessSuccesses          int            `json:"pageProcessSuccesses"`
	PageProcessSuccessesThisRound int            `json:"pageProcessSuccessesThisRound"`
	MaxRounds                     int            `json:"maxRounds"`
	Repeat                        float64        `json:"repeat"`
	CrawlDelay                    float64        `json:"crawlDelay"`
	ObeyRobots                    int            `json:"obeyRobots"`
	MaxToCrawl                    int            `json:"maxToCrawl"`
	MaxToProcess                  int            `json:"maxToProcess"`
	OnlyProcessIfNew              int            `json:"onlyProcessIfNew"`
	Seeds                         string         `json:"seeds"`
	RoundsCompleted               int            `json:"roundsCompleted"`
	RoundStartTime                int            `json:"roundStartTime"`
	CurrentTime                   int            `json:"currentTime"`
	CurrentTimeUTC                int            `json:"currentTimeUTC"`
	ApiUrl                        string         `json:"apiUrl"`
	UrlCrawlPattern               string         `json:"urlCrawlPattern"`
	UrlProcessPattern             string         `json:"urlProcessPattern"`
	PageProcessPattern            string         `json:"pageProcessPattern"`
	UrlCrawlRegEx                 string         `json:"urlCrawlRegEx"`
	UrlProcessRegEx               string         `json:"urlProcessRegEx"`
	MaxHops                       int            `json:"maxHops"`
	DownloadJson                  string         `json:"downloadJson"`
	DownloadUrls                  string         `json:"downloadUrls"`
	NotifyEmail                   string         `json:"notifyEmail"`
	NotifyWebhook                 string         `json:"notifyWebhook"`
}

type jobStatusType struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

//
// The Crawlbot API allows you to programmatically manage Crawlbot crawls and retrieve output. Crawlbot API
// responses are in JSON format.
//
// Creating or Updating a Crawl
//
// To create a crawl, make a GET request to http://api.diffbot.com/v3/crawl.
//
// Provide the following data:
//
//	+-----------+-------------------------------------------------------------------------+
//	| PARAMETER | DESCRIPTION                                                             |
//	+-----------+-------------------------------------------------------------------------+
//	| token     | Developer token                                                         |
//	| name      | Job name. This should be a unique identifier and can be used to modify  |
//	|           | your crawl or retrieve its output.                                      |
//	| seeds     | Seed URL(s). Must be URL encoded. Separate multiple URLs with           |
//	|           | whitespace to spider multiple sites within the same crawl. By default   |
//	|           | Crawlbot will restrict spidering to the entire domain                   |
//	|           | ("http://blog.diffbot.com" will include URLs at                         |
//	|           | "http://www.diffbot.com").											  |
//	| apiUrl    | Full Diffbot API URL through which to process pages. E.g.,              |
//	|           | &apiUrl=http://api.diffbot.com/v3/article to process matching links via |
//	|           | the Article API. The Diffbot API URL can include querystring parameters |
//	|           | to tailor the output. For example,                                      |
//	|           | &apiUrl=http://api.diffbot.com/v3/product?fields=querystring,meta will  |
//	|           | process matching links using the Product API, and also return the       |
//	|           | querystring and meta fields.                                            |
//	+-----------+-------------------------------------------------------------------------+
//
// To automatically identify and process content using our Page Classifier API (Smart Processing), pass
// apiUrl=http://api.diffbot.com/v3/analyze?mode=auto to return all page-types. See full Page Classifier
// documentation under the Automatic APIs documentation.
//
// Be sure to URL encode your Diffbot API actions.
// You can refine your crawl using the following optional controls. Read more on crawling versus processing.
//
//	+-------------------+-------------------------------------------------------------------------+
//	| PARAMETER         | DESCRIPTION                                                             |
//	+-------------------+-------------------------------------------------------------------------+
//	| urlCrawlPattern   | Specify ||-separated strings to limit pages crawled to those whose URLs |
//	|                   | contain any of the content strings. You can use the exclamation point to|
//	|                   | specify a negative string, e.g. !product to exclude URLs containing the |
//	|                   | string "product," and the ^ and $ characters to limit matches to the    |
//	|                   | beginning or end of the URL.                                            |
//	|                   | The use of a urlCrawlPattern will allow Crawlbot to spider outside of   |
//	|                   | the seed domain; it will follow all matching URLs regardless of domain. |
//	| urlCrawlRegEx	    | Specify a regular expression to limit pages crawled to those URLs that  |
//	|                   | match your expression. This will override any urlCrawlPattern value.    |
//	|                   | The use of a urlCrawlRegEx will allow Crawlbot to spider outside of the |
//	|                   | seed domain; it will follow all matching URLs regardless of domain.     |
//	| urlProcessPattern | Specify ||-separated strings to limit pages processed to those whose    |
//	|                   | URLs contain any of the content strings. You can use the exclamation    |
//	|                   | point to specify a negative string, e.g. !/category to exclude URLs     |
//	|                   | containing the string "/category," and the ^ and $ characters to limit  |
//	|                   | matches to the beginning or end of the URL.                             |
//	| urlProcessRegEx   | Specify a regular expression to limit pages processed to those URLs that|
//	|                   | match your expression. This will override any urlProcessPattern value.  |
//	| pageProcessPattern| Specify ||-separated strings to limit pages processed to those whose    |
//	|                   | HTML contains any of the content strings.                               |
//	+-------------------+-------------------------------------------------------------------------+
//	| Additional (optional) Parameters:                                                           |
//	+-------------------+-------------------------------------------------------------------------+
//	| maxHops           | Specify the depth of your crawl. A maxHops=0 will limit processing to   |
//	|                   | the seed URL(s) only -- no other links will be processed; maxHops=1 will|
//	|                   | process all (otherwise matching) pages whose links appear on seed       |
//	|                   | URL(s); maxHops=2 will process pages whose links appear on those pages; |
//	|                   | and so on.                                                              |
//	|                   | By default (maxHops=-1) Crawlbot will crawl and process links at any    |
//	|                   | depth.                                                                  |
//	| maxToCrawl        | Specify max pages to spider. Default: 100,000.                          |
//	| maxToProcess      | Specify max pages to process through Diffbot APIs. Default: 100,000.    |
//	| notifyEmail       | Send a message to this email address when the crawl hits the maxToCrawl |
//	|                   | or maxToProcess limit, or when the crawl completes.                     |
//	| notifyWebhook	    | Pass a URL to be notified when the crawl hits the maxToCrawl or         |
//	|                   | maxToProcess limit, or when the crawl completes. You will receive a POST|
//	|                   | with X-Crawl-Name and X-Crawl-Status in the headers, and the full JSON  |
//	|                   | response in the POST body.                                              |
//	|                   | We've integrated with Zapier to make webhooks even more powerful; read  |
//	|                   | more on what you can do with Zapier and Diffbot.                        |
//	| crawlDelay        | Wait this many seconds between each URL crawled from a single IP        |
//	|                   | address. Specify the number of seconds as an integer or floating-point  |
//	|                   | number (e.g., crawlDelay=0.25).                                         |
//	| repeat            | Specify the number of days as a floating-point (e.g. repeat=7.0) to     |
//	|                   | repeat this crawl. By default crawls will not be repeated.              |
//	| onlyProcessIfNew  | By default repeat crawls will only process new (previously unprocessed) |
//	|                   | pages. Set to 0 (onlyProcessIfNew=0) to process all content on repeat   |
//	|                   | crawls.                                                                 |
//	| maxRounds	Specify | the maximum number of crawl repeats. By default (maxRounds=0) repeating |
//	|                   | crawls will continue indefinitely.                                      |
//	+-------------------+-------------------------------------------------------------------------+
//
// Response
//
// Upon adding a new crawl, you will receive a success message in the JSON response, in addition to
// full crawl details:
//
//  "response": "Successfully added urls for spidering."
//
//
// Pausing, Restarting or Deleting Crawls
//
// You can manage your crawls by making GET requests to the same endpoint, http://api.diffbot.com/v3/crawl.
//
// Provide the following data:
//
//	+------------+-------------------------------------------------------------------------+
//	| PARAMETER  | DESCRIPTION                                                             |
//	+------------+-------------------------------------------------------------------------+
//	| token      | Developer token                                                         |
//	| name       | Job name as defined when the crawl was created.                         |
//	+------------+-------------------------------------------------------------------------+
//	| Job-control arguments                                                                |
//	+------------+-------------------------------------------------------------------------+
//	| roundStart | Pass roundStart=1 to force the start of a new crawl "round" (manually   |
//	|            | repeat the crawl). If onlyProcessIfNew is set to 1 (default), only      |
//	|            | newly-created pages will be processed.                                  |
//	| pause      | Pass pause=1 to pause a crawl. Pass pause=0 to resume a paused crawl.   |
//	| restart    | Restart removes all crawled data while maintaining crawl settings. Pass |
//	|            | restart=1 to restart a crawl.                                           |
//	| delete     | Pass delete=1 to delete a crawl, and all associated data, completely.   |
//	+------------+-------------------------------------------------------------------------+
//
// Retrieving Crawlbot API Data
//
// To download results please make a GET request to http://api.diffbot.com/v3/crawl/data. Provide the following
// arguments based on the data you need. By default the complete extracted JSON data will be downloaded.
//
//	+-----------+-------------------------------------------------------------------------+
//	| PARAMETER | DESCRIPTION                                                             |
//	+-----------+-------------------------------------------------------------------------+
//	| token	    | Diffbot token.                                                          |
//	| name      | Name of the crawl whose data you wish to download.                      |
//	| format    | Request format=csv to download the extracted data in CSV format         |
//	|           | (default: json). Note that CSV files will only contain top-level fields.|
//	+-----------+-------------------------------------------------------------------------+
//	| For diagnostic data:                                                                |
//	+-----------+-------------------------------------------------------------------------+
//	| type      | Request type=urls to retrieve the crawl URL Report (CSV).               |
//	| num       | Pass an integer value (e.g. num=100) to request a subset of URLs, most  |
//	|           | recently crawled first.                                                 |
//	+-----------+-------------------------------------------------------------------------+
//
// Viewing Crawl Details
//
// Your active crawls (and any active Bulk API jobs) will be returned in the jobs object in a request made to
// http://api.diffbot.com/v3/crawl.
//
// To retrieve a single crawl's details, provide the crawl's name in your request:
//
//	+-----------+-------------------------------------------------------------------------+
//	| PARAMETER | DESCRIPTION                                                             |
//	+-----------+-------------------------------------------------------------------------+
//	| token     | Developer token                                                         |
//	| name      | Name of crawl to retrieve.                                              |
//	+-----------+-------------------------------------------------------------------------+
//
// To view all crawls, simply omit the name parameter.
//
// Response
//
// This will return a JSON response of your token's crawls and Bulk API jobs. Sample response from a single crawl:
//
//  {
//    "jobs": [
//      {
//        "name": "crawlJob",
//        "type": "crawl",
//        "jobCreationTimeUTC": 1427410692,
//        "jobCompletionTimeUTC": 1427410798,
//        "jobStatus": {
//          "status": 9,
//          "message": "Job has completed and no repeat is scheduled."
//        },
//        "sentJobDoneNotification": 1,
//        "objectsFound": 177,
//        "urlsHarvested": 2152,
//        "pageCrawlAttempts": 367,
//        "pageCrawlSuccesses": 365,
//        "pageCrawlSuccessesThisRound": 365,
//        "pageProcessAttempts": 210,
//        "pageProcessSuccesses": 210,
//        "pageProcessSuccessesThisRound": 210,
//        "maxRounds": 0,
//        "repeat": 0.0,
//        "crawlDelay": 0.25,
//        "obeyRobots": 1,
//        "maxToCrawl": 100000,
//        "maxToProcess": 100000,
//        "onlyProcessIfNew": 1,
//        "seeds": "http://support.diffbot.com",
//        "roundsCompleted": 0,
//        "roundStartTime": 0,
//        "currentTime": 1443822683,
//        "currentTimeUTC": 1443822683,
//        "apiUrl": "http://api.diffbot.com/v3/analyze",
//        "urlCrawlPattern": "",
//        "urlProcessPattern": "",
//        "pageProcessPattern": "",
//        "urlCrawlRegEx": "",
//        "urlProcessRegEx": "",
//        "maxHops": -1,
//        "downloadJson": "http://api.diffbot.com/v3/crawl/download/sampletoken-crawlJob_data.json",
//        "downloadUrls": "http://api.diffbot.com/v3/crawl/download/sampletoken-crawlJob_urls.csv",
//        "notifyEmail": "support@diffbot.com",
//        "notifyWebhook": "http://www.diffbot.com"
//      }
//    ]
//  }
//
// Status Codes
//
// The jobStatus object will return the following status codes and associated messages:
//
//	+-----------+-------------------------------------------------------------------------+
//	| STATUS    | DESCRIPTION                                                             |
//	+-----------+-------------------------------------------------------------------------+
//	| 0         | Job is initializing                                                     |
//	| 1         | Job has reached maxRounds limit                                         |
//	| 2         | Job has reached maxToCrawl limit                                        |
//	| 3         | Job has reached maxToProcess limit                                      |
//	| 4         | Next round to start in _____ seconds                                    |
//	| 5         | No URLs were added to the crawl                                         |
//	| 6         | Job paused                                                              |
//	| 7         | Job in progress                                                         |
//	| 8         | All crawling temporarily paused by root administrator for maintenance.  |
//	| 9         | Job has completed and no repeat is scheduled                            |
//	+-----------+-------------------------------------------------------------------------+
//

type CrawlResponse struct {
	Response string   `json:"response,omitempty"`
	Jobs     []*Crawl `json:"jobs,omitempty"`
}

func CreateCrawl(client *http.Client, token, name string, seeds []string, apiUrl string, opt *Options) (*CrawlResponse, error) {
	values := url.Values{}
	for _, seed := range seeds {
		values.Add("seeds", seed)
	}
	values.Add("apiUrl", apiUrl)

	body, err := Crawlbot(client, "crawl", token, name, values, opt)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var result CrawlResponse
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// method: ["roundStart", "pause", "resume", "restart", "delete"]
func EditCrawl(client *http.Client, token, name, method string) (*CrawlResponse, error) {
	if method == "" {
		return nil, fmt.Errorf("No method specified")
	}

	values := url.Values{}
	switch method {
	case "roundStart":
		values.Add("roundStart", "1")
	case "pause":
		values.Add("pause", "1")
	case "resume":
		values.Add("pause", "0")
	case "restart":
		values.Add("restart", "1")
	case "delete":
		values.Add("delete", "1")
	}

	body, err := Crawlbot(client, "crawl", token, name, values, nil)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var result CrawlResponse
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func ViewCrawl(client *http.Client, token, name string, opt *Options) (*CrawlResponse, error) {
	values := url.Values{}
	body, err := Crawlbot(client, "crawl", token, name, values, opt)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var result CrawlResponse
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (p *CrawlResponse) String() string {
	d, _ := json.Marshal(p)
	return string(d)
}

type CrawlData []*Classification

// retrieves the details of the crawl
func RetrieveCrawl(client *http.Client, token, name string, opt *Options) ([]*Classification, error) {
	values := url.Values{}
	body, err := Crawlbot(client, "crawl/data", token, name, values, opt)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	results := make([]*Classification, 0)
	reader := bufio.NewReader(body)

	i := 0
	for {
		// skip the first line
		b, err := reader.ReadBytes('\n')
		if i == 0 {
			b, err = reader.ReadBytes('\n')
		}
		if err == io.EOF {
			return results, nil
		}
		if err != nil {
			break
		}

		var result *Classification
		err = json.Unmarshal(b, &result)
		if err != nil {
			continue
		}

		results = append(results, result)
	}

	return results, err
}

func (p *CrawlData) String() string {
	d, _ := json.Marshal(p)
	return string(d)
}

func Crawlbot(client *http.Client, method, token, name string, params url.Values, opt *Options) (io.ReadCloser, error) {
	return CrawlbotServer(client, DefaultServer, method, token, name, params, opt)
}

func CrawlbotServer(client *http.Client, server, method, token, name string, params url.Values, opt *Options) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", makeCrawlRequestUrl(server, method, token, name, params, opt), nil)
	if err != nil {
		return nil, err
	}
	if opt != nil && opt.CustomHeader != nil {
		req.Header = opt.CustomHeader
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func makeCrawlRequestUrl(server, method, token, name string, params url.Values, opt *Options) string {
	return fmt.Sprintf("%s/%s?token=%s&name=%s&%s&%s",
		server, method, token, name, params.Encode(), opt.MethodParamString(method).Encode(),
	)
}
