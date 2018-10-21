// Copyright 2014 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package diffbot implements a Diffbot client library.

Diffbot using AI, computer vision, machine learning and
natural language processing, Diffbot provides developers
numerous tools to understand and extract from any web page.

Generic API

The basic API is diffbot.DiffbotServer:

	import (
		"github.com/diffbot/diffbot-go-client"
	)

	var (
		token = `0123456789abcdef0123456789abcdef` // invalid token, just a example
		url   = `http://blog.diffbot.com/diffbots-new-product-api-teaches-robots-to-shop-online/`
	)

	func main() {
		respBody, err := diffbot.DiffbotServer(diffbot.DefaultServer, "article", token, url, nil)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(respBody))
	}

The diffbot.Diffbot API use the diffbot.DefaultServer as the server.

Article API

Tha Article API use the diffbot.Diffbot to invoke the "article" method,
and convert the reponse body to diffbot.Article struct.

	func main() {
		article, err := diffbot.ParseArticle(token, url, nil)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(article)
	}

Frontpage API

Tha Frontpage API use the diffbot.Diffbot to invoke the "frontpage" method,
and convert the reponse body to diffbot.Frontpage struct.

	func main() {
		page, err := diffbot.ParseFrontpage(token, url, nil)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(page)
	}

Image API

Tha Image API use the diffbot.Diffbot to invoke the "image" method,
and convert the reponse body to diffbot.Image struct.

	func main() {
		imgInfo, err := diffbot.ParseImage(token, url, nil)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(imgInfo)
	}

Product API

Tha Product API use the diffbot.Diffbot to invoke the "product" method,
and convert the reponse body to diffbot.Product struct.

	func main() {
		product, err := diffbot.ParseProduct(token, url, nil)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(product)
	}

Classification API

Tha Classification API use the diffbot.Diffbot to invoke the "analyze" method,
and convert the reponse body to diffbot.Classification struct.

	func main() {
		info, err := diffbot.ParseClassification(token, url, nil)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(info)
	}

Options

We use `diffbot.Options` to specify the options:

	func main() {
		opt := &diffbot.Options{
			Fields:   "meta,querystring,images(*)",
			Timeout:  time.Millisecond * 1000,
			Callback: "",
		}
		fmt.Println(opt.MethodParamString("article"))
		// Output:
		// &fields=meta,querystring,images(*)&timeout=1000
	}

You can call Diffbot with custom headers:

	func main() {
		opt := diffbot.Options{}
		opt.CustomHeader.Add("X-Forward-Cookie", "abc=123")
		respBody, err := diffbot.Diffbot(token, url, opt)
		...
	}

Error handling

If diffbot server return error message, it will be converted to the `diffbot.Error`:

	func main() {
		respBody, err := diffbot.Diffbot(token, url, nil)
		if err := nil {
			if apiErr, ok := err.(*diffbot.ApiError); ok {
				// ApiError, e.g. {"error":"Not authorized API token.","errorCode":401}
			}
			log.Fatal(err)
		}
		fmt.Println(string(respBody))
	}

Other

Diffbot API Document at http://diffbot.com/dev/docs/ or http://diffbot.com/products/.

Please report bugs to <chaishushan@gmail.com>.

*/
package diffbot
