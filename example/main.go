package main

import (
	"log"
	"net"
	"net/http"

	_ "net/http/pprof"

	"github.com/kavehmz/short"
)

func main() {
	site := short.Site{Host: "https://short.kaveh.me/"}
	http.HandleFunc("/", site.Redirect)
	http.HandleFunc("/post", site.Post)

	// If pool is full, connections will wait.
	// This is not a good pattern for high scale sites.
	// This only helps if http connection as a resource is cheaper
	// than underlying resources like db connetion,...
	maxServingClients := 2
	maxClientsPool := make(chan bool, maxServingClients)

	server := &http.Server{
		Addr:    ":8080",
		Handler: nil,
		ConnState: func(conn net.Conn, state http.ConnState) {
			switch state {
			case http.StateNew:
				maxClientsPool <- true
			case http.StateClosed, http.StateHijacked:
				<-maxClientsPool

			}
		},
	}
	log.Fatal(server.ListenAndServe())
}
