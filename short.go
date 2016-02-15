package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/asaskevich/govalidator"
	"github.com/garyburd/redigo/redis"
)

func redisdb() redis.Conn {
	redisdb, err := redis.DialURL(os.Getenv("REDISURL"))
	if err != nil {
		panic(err)
	}
	return redisdb
}

func saveShort(url string) (shortest string) {
	redisdb := redisdb()
	defer redisdb.Close()

	hash := fmt.Sprintf("%x", md5.Sum([]byte(url)))

	old, _ := redis.String(redisdb.Do("GET", "i:"+hash))
	if old != "" {
		return old
	}

	// Finding the shortest hash which does not exist.
	for i := 1; i <= 32; i++ {
		s, _ := redisdb.Do("GET", hash[0:i])
		if s == nil {
			shortest = hash[0:i]
			break
		}
	}
	redisdb.Do("SET", shortest, url)
	redisdb.Do("SET", "i:"+hash, shortest)
	return shortest
}

type shortRequest struct {
	URL string
}

func post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	decoder := json.NewDecoder(r.Body)
	var t shortRequest
	decoder.Decode(&t)

	if !govalidator.IsURL(t.URL) {
		fmt.Fprintf(w, "{\"error\":\"invalid url\"}")
		return
	}
	short := saveShort(t.URL)
	if short == "" {
		fmt.Fprintf(w, "{\"error\":\"url shortening failed\"}")
		return
	}
	fmt.Fprintf(w, "{\"short\":\"https://short.kaveh.me/%s\"}", short)
}

func redirect(w http.ResponseWriter, r *http.Request) {
	redisdb := redisdb()
	defer redisdb.Close()

	u, _ := redis.String(redisdb.Do("GET", r.URL.Path[1:]))
	if u == "" {
		if r.Header.Get("Content-Type") == "application/json" {
			w.Header().Set("Content-Type", "application/javascript")
			fmt.Fprintf(w, "{\"error\":\"not found\"}")
			return
		}
		fmt.Fprintf(w, "not found")
		return
	}
	if r.Header.Get("Content-Type") == "application/json" {
		w.Header().Set("Content-Type", "application/javascript")
		fmt.Fprintf(w, "{\"url\":\"%s\"}", u)
		return
	}
	http.Redirect(w, r, u, http.StatusMovedPermanently)
}

func main() {
	http.HandleFunc("/", redirect)
	http.HandleFunc("/post", post)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
