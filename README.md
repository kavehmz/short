# short
A sample URL shortener Golang app using Redis

[![Go Lang](http://kavehmz.github.io/static/gopher/gopher-front.svg)](https://golang.org/)
[![Build Status](https://travis-ci.org/kmzarc/short.svg?branch=master)](https://travis-ci.org/kmzarch/short)
[![Coverage Status](https://coveralls.io/repos/github/kmzarc/short/badge.svg?branch=master)](https://coveralls.io/github/kmzarc/short?branch=master)

---
### short
short is a simple web based api. It will produce shortest hash possible for a url. It will use longer hashes only in case of collision.

short has has two entries:

/post which accepts a url and returns a short version of it.

{"url": "https://example.org"}

/TOKEN which based on request headers will return a json reply or will redirect the agent to the original url.


### How to check

You need to have a usable go installation. Then create a directory named site and follow the commands:

```
$ cd site
$ GOPATH=$PWD
$ go get -u github.com/kmzarc/short
```

This will install the project and all its dependencies.

Now to test it create a redis instance in redislabs.com and export you redis url like:

```$ REDISURL="redis://:PASSWORD@pub-zone.redislabs.com:12919/0"```

Then you can run the example

```$ go run src/github.com/kmzarc/short/example/main.go```

To test it use curl

```
$ # To send a url to be minimized
$ curl -v -X POST -H 'Content-Type: application/json' -d '{"url":"https://example.org"}' http://localhost:8080/post
$ # json query
$ curl -v -X POST -H 'Content-Type: application/json'  http://localhost:8080/5
$ # normal redirect
$ curl http://localhost:8080/5
```
