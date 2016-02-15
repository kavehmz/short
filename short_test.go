package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestIndex(t *testing.T) {
	req, err := http.NewRequest("GET", "http://short.me/", nil)
	errCheck(err)
	w := httptest.NewRecorder()
	index(w, req)
	if w.Code != 200 {
		t.Error("expected code was 200 but we got ", w.Code)
	}
}

func TestPost(t *testing.T) {
	data := url.Values{}
	data.Add("url", "https://example.org")

	req, _ := http.NewRequest("POST", "http://short.me/post", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	w := httptest.NewRecorder()
	post(w, req)
	if w.Code != 200 {
		t.Error("expected code was 200 but we got ", w.Code)
	}
	if w.Body.String() != "http://short.me/5" {
		t.Error("produced value is no correct", w.Code)
	}
}
