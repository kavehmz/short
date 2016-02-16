/*
Package short is a simple web based api. It will produce shortest hash possible for a url. It will use longer hashes only in case of collision.

For example for https://example.org is redis is empty it will return /5

short has has two entries:

/post which accepts a url and returns a short version of it.

	{"url": "https://example.org"}

/TOKEN which based on request headers will return a json reply or will redirect the agent to the original url.

short uses redis for its backned storage. You can set the redis url in command or in code.

	$ REDISURL="redis://:PASSWORD@pub-zone.redislabs.com:12919/0"

Example of using this package is available in exmaple directory.
*/
package short

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/asaskevich/govalidator"
	"github.com/garyburd/redigo/redis"
)

//Site defines main attributes of shortener service.
type Site struct {
	// Host will define the prefix to be used for generated short token for redirects.
	Host string
	// RedisURL defines the redis instance to use. If it is empty REDISURL environment variable will be used.
	RedisURL string
}

func (site Site) redisURL() string {
	if site.RedisURL != "" {
		return site.RedisURL
	}
	return os.Getenv("REDISURL")
}

func (site Site) redisdb() redis.Conn {
	redisdb, err := redis.DialURL(site.redisURL())
	if err != nil {
		panic(err)
	}
	return redisdb
}

func (site Site) saveShort(url string) (shortest string, err error) {
	if !govalidator.IsURL(url) {
		return "", errors.New("invalid url")
	}

	redisdb := site.redisdb()
	defer redisdb.Close()

	hash := fmt.Sprintf("%x", md5.Sum([]byte(url)))

	old, _ := redis.String(redisdb.Do("GET", "i:"+hash))
	if old != "" {
		return site.Host + old, nil
	}

	// Finding the shortest hash which does not exist.
	for i := 1; i <= 32; i++ {
		s, _ := redisdb.Do("GET", hash[0:i])
		if s == nil {
			shortest = hash[0:i]
			break
		}
	}
	if shortest == "" {
		return "", errors.New("url shortening failed")
	}

	redisdb.Do("SET", shortest, url)
	redisdb.Do("SET", "i:"+hash, shortest)
	return site.Host + shortest, nil
}

type shortRequest struct {
	URL string
}

/*
Post Only accepts json data in the following format

	{"url": "https://example.org"}

return results is always in the following format:

	{"short":"https://short.me/ab", "error":""}

If there was any error short will be empty and error message will show.
*/
func (site Site) Post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	decoder := json.NewDecoder(r.Body)
	var t shortRequest
	decoder.Decode(&t)

	shortURL, err := site.saveShort(t.URL)
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	if r.Header.Get("Content-Type") == "application/json" {
		fmt.Fprintf(w, "{\"short\":\"%s\", \"error\":\"%s\"}", shortURL, errMsg)
		return
	}
}

/*
Redirect will either return a json value in the following format:

 	{"url":"https://example.org", "error":""}

or it will do a 301, http redirect. Return action is based on headers.
*/
func (site Site) Redirect(w http.ResponseWriter, r *http.Request) {
	redisdb := site.redisdb()
	defer redisdb.Close()

	t, _ := redis.String(redisdb.Do("GET", r.URL.Path[1:]))
	u, _ := url.Parse(t)

	errMsg := ""
	if u.String() == "" {
		errMsg = "not found"
	}

	if r.Header.Get("Content-Type") == "application/json" {
		w.Header().Set("Content-Type", "application/javascript")
		fmt.Fprintf(w, "{\"url\":\"%s\", \"error\":\"%s\"}", u.String(), errMsg)
		return
	}

	if errMsg != "" {
		fmt.Fprintf(w, errMsg)
		return
	}

	http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
}
