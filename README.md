# short
A sample URL shortener Golang app using Redis

[![Go Lang](http://kavehmz.github.io/static/gopher/gopher-front.svg)](https://golang.org/)
[![GoDoc](https://godoc.org/github.com/kavehmz/short?status.svg)](https://godoc.org/github.com/kavehmz/short)
[![Build Status](https://travis-ci.org/kavehmz/short.svg?branch=master)](https://travis-ci.org/kavehmz/short)
[![Coverage Status](https://coveralls.io/repos/github/kavehmz/short/badge.svg?branch=master)](https://coveralls.io/github/kavehmz/short?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/kavehmz/short)](https://goreportcard.com/report/github.com/kavehmz/short)

---
### short
short is a simple web based api. It will produce shortest hash possible for a url. It will use longer hashes only in case of collision.

For example for https://example.org is redis is empty it will return /5

short has has two entries:

/post which accepts a url and returns a short version of it.

{"url": "https://example.org"}

/TOKEN which based on request headers will return a json reply or will redirect the agent to the original url.

### Usage

```go
package main

import (
	"log"
	"net/http"

	"github.com/kavehmz/short"
)

func main() {
	site := short.Site{Host: "https://short.kaveh.me/"}
	http.HandleFunc("/", site.Redirect)
	http.HandleFunc("/post", site.Post)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

```

If you like to see a sample implementation, which uses Heroku and limits maximum number of clients take a look at "cmd/shortsite/main.go" in:

https://github.com/kmzarc/shortsite

### How to check

You need to have a usable go installation. Then create a directory named site and follow the commands:

```
$ cd site
$ GOPATH=$PWD
$ go get -u github.com/kavehmz/short
```

This will install the project and all its dependencies.

Now to test it create a redis instance in redislabs.com and export your redis url like:

```$ REDISURL="redis://:PASSWORD@pub-zone.redislabs.com:12919/0"```

Then you can run the example

```$ go run src/github.com/kavehmz/short/example/main.go```

To test it use curl

```
$ # To send a url to be minimized
$ curl -v -X POST -H 'Content-Type: application/json' -d '{"url":"https://example.org"}' http://localhost:8080/post
$ # json query
$ curl -v -X GET -H 'Content-Type: application/json'  http://localhost:8080/5
$ # normal redirect
$ curl http://localhost:8080/5
```
