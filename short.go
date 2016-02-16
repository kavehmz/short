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

//Site is good
type Site struct {
	Host     string
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

// Post Only accepts JSON
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

// Redirect Only accepts JSON
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
