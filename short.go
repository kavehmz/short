package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/asaskevich/govalidator"
	"github.com/garyburd/redigo/redis"
)

func errCheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func redisdb() redis.Conn {
	redisdb, err := redis.DialURL(os.Getenv("REDISURL"))
	errCheck(err)
	return redisdb
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path[1:] != "" {
		redirect(w, r)
		return
	}
	t, err := template.ParseFiles(os.Getenv("GOPATH") + "/static/index.html")
	errCheck(err)
	t.Execute(w, nil)
}

func saveShort(url string) (string, error) {
	redisdb := redisdb()
	defer redisdb.Close()

	short := fmt.Sprintf("%x", md5.Sum([]byte(url)))

	old, _ := redis.String(redisdb.Do("GET", short))
	if old != "" {
		return old, nil
	}

	for i := 1; i <= 32; i++ {
		s, err := redisdb.Do("GET", short[0:i])
		errCheck(err)
		if s == nil {
			_, err = redisdb.Do("SET", short[0:i], url)
			_, err = redisdb.Do("SET", short, short[0:i])
			errCheck(err)
			return short[0:i], nil
		}
	}
	return "", errors.New("Can't find a uniqie key available.")

}

func post(w http.ResponseWriter, r *http.Request) {
	long := r.FormValue("url")
	if !govalidator.IsURL(long) {
		fmt.Fprintf(w, "Invalid url: %s", long)
		return
	}
	short, err := saveShort(long)
	errCheck(err)
	fmt.Fprintf(w, r.URL.Scheme+"://"+r.URL.Host+"/%s", short)
}

func redirect(w http.ResponseWriter, r *http.Request) {
	redisdb := redisdb()
	defer redisdb.Close()

	u, _ := redis.String(redisdb.Do("GET", r.URL.Path[1:]))
	if u == "" {
		fmt.Fprintf(w, "not found")
		return
	}
	http.Redirect(w, r, u, http.StatusMovedPermanently)
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/post", post)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
