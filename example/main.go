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
