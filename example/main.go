package main

import (
	"log"
	"net/http"

	"github.com/kmzarc/short"
)

func main() {
	site := short.Site{Host: "https://localhost:8080/", Port: 8080}
	http.HandleFunc("/", site.Redirect)
	http.HandleFunc("/post", site.Post)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
