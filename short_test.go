package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestPost(t *testing.T) {
	redisdb().Do("FLUSHALL")
	// Runing twice must still return the same short url '/5'
	for i := 0; i < 2; i++ {
		var jsonStr = []byte(`{"url":"https://example.org"}`)
		req, _ := http.NewRequest("POST", "http://short.me/post", bytes.NewBuffer(jsonStr))
		req.Header.Add("Content-Type", "application/json")

		w := httptest.NewRecorder()
		post(w, req)
		if w.Code != 200 {
			t.Error("expected code was 200 but we got ", w.Code)
		}
		if w.Body.String() != "{\"short\":\"https://short.kaveh.me/5\"}" {
			t.Error("produced value is not correct", w.Body.String())
		}
	}
}

func TestPostError(t *testing.T) {
	redisdb := redisdb()
	redisdb.Do("FLUSHALL")
	short := "5564fd6a95028f02e52b38bb1743c816"
	// Add all possible lenght of md5 to redis and see if post correctly fails
	for i := 1; i <= 32; i++ {
		redisdb.Do("SET", short[0:i], "https://example.org")
	}

	var jsonStr = []byte(`{"url":"https://example.org"}`)
	req, _ := http.NewRequest("POST", "http://short.me/post", bytes.NewBuffer(jsonStr))
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()
	post(w, req)
	if w.Body.String() != "{\"error\":\"url shortening failed\"}" {
		fmt.Println(w.Body.String())
		t.Error("produced value is not correct", w.Body.String())
	}
}

func TestPostURLError(t *testing.T) {
	var jsonStr = []byte(`{"url":"https//ex ample.org"}`)
	req, _ := http.NewRequest("POST", "http://short.me/post", bytes.NewBuffer(jsonStr))
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()
	post(w, req)
	if w.Body.String() != "{\"error\":\"invalid url\"}" {
		fmt.Println(w.Body.String())
		t.Error("produced value is not correct", w.Body.String())
	}
}

func TestRedirect(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://short.kaveh.me/5", nil)
	w := httptest.NewRecorder()
	redirect(w, req)
	if w.Code != 301 {
		t.Error("expected code was 301 but we got ", w.Code)
	}
}

func TestRedirectNotFoudn(t *testing.T) {
	redisdb().Do("FLUSHALL")
	req, _ := http.NewRequest("GET", "http://short.kaveh.me/5", nil)
	w := httptest.NewRecorder()
	redirect(w, req)
	if w.Body.String() != "not found" {
		t.Error("unknown hash did not return any error")
	}
}

func TestRedisError(t *testing.T) {
	os.Setenv("REDISURL", "")
	defer func() {
		recover()
	}()
	redisdb()

	t.Error("wrong REDISURL didn't cause any error")
}

func TestMain(t *testing.T) {
	go main()
	time.Sleep(time.Second)
}
