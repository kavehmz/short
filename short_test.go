package short

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestPost(t *testing.T) {
	site := Site{Host: "https://short.me/"}
	site.redisdb().Do("FLUSHALL")
	var jsonStr = []byte(`{"url":"https://example.org"}`)
	req, _ := http.NewRequest("POST", "https://short.me/post", bytes.NewBuffer(jsonStr))
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()
	site.Post(w, req)
	if w.Code != 200 {
		t.Error("expected code was 200 but we got ", w.Code)
	}
	if w.Body.String() != "{\"short\":\"https://short.me/5\", \"error\":\"\"}" {
		t.Error("produced value is not correct", w.Body.String())
	}

}

func TestPostError(t *testing.T) {
	site := Site{Host: "https://short.me/"}
	site.redisdb().Do("FLUSHALL")
	var jsonStr = []byte(`{"url":"https://examp le.org"}`)
	req, _ := http.NewRequest("POST", "https://short.me/post", bytes.NewBuffer(jsonStr))
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()
	site.Post(w, req)
	if w.Code != 200 {
		t.Error("expected code was 200 but we got ", w.Code)
	}
	if w.Body.String() != "{\"short\":\"\", \"error\":\"invalid url\"}" {
		t.Error("produced value is not correct", w.Body.String())
	}

}

func TestSaveShort(t *testing.T) {
	site := Site{Host: "https://short.me/"}
	site.redisdb().Do("FLUSHALL")
	// Running twice must still return the same short url '/5'
	u, err := site.saveShort("https://example.org")
	if u != "https://short.me/5" || err != nil {
		t.Error("produced value is not correct", u)
	}
	u, err = site.saveShort("https://example.org")
	if u != "https://short.me/5" || err != nil {
		t.Error("produced value is not the same", u)
	}
	u, err = site.saveShort("https://examp e.org")
	if u != "" || err == nil {
		t.Error("wrong url did not cause error", u)
	}

	site.redisdb().Do("FLUSHALL")
	s := "5564fd6a95028f02e52b38bb1743c816"
	redisdb := site.redisdb()
	// Add all possible length of md5 to redis and see if post correctly fails
	for i := 1; i <= 32; i++ {
		redisdb.Do("SET", s[0:i], "1")
	}
	u, err = site.saveShort("https://example.org")
	if u != "" || err == nil {
		t.Error("did not cause issue when all slots are full", u)
	}
}

func TestRedirect(t *testing.T) {
	site := Site{Host: "https://short.me/"}
	site.redisdb().Do("FLUSHALL")
	site.saveShort("https://example.org")
	req, _ := http.NewRequest("GET", "https://short.me/5", nil)
	w := httptest.NewRecorder()
	site.Redirect(w, req)
	if w.Code != 301 {
		t.Error("expected code was 301 but we got ", w.Code)
	}
}

func TestRedirectNotFound(t *testing.T) {
	site := Site{Host: "https://short.me/"}
	site.redisdb().Do("FLUSHALL")
	req, _ := http.NewRequest("GET", "https://short.me/5", nil)
	w := httptest.NewRecorder()
	site.Redirect(w, req)
	if w.Body.String() != "not found" {
		t.Error("unknown hash did not return any error")
	}
}

func TestRedirectJson(t *testing.T) {
	site := Site{Host: "https://short.me"}
	site.redisdb().Do("FLUSHALL")
	site.saveShort("https://example.org")
	req, _ := http.NewRequest("GET", "https://short.me/5", nil)
	req.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	site.Redirect(w, req)
	if w.Body.String() != "{\"url\":\"https://example.org\", \"error\":\"\"}" {
		t.Error("produced value is not correct", w.Body.String())
	}
}

func TestRedirectJsonNotFound(t *testing.T) {
	site := Site{Host: "https://short.me/"}
	req, _ := http.NewRequest("GET", "http://kaveh.me/xx", nil)
	req.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	site.Redirect(w, req)
	if w.Body.String() != "{\"url\":\"\", \"error\":\"not found\"}" {
		t.Error("produced value is not correct", w.Body.String())
	}
}

func TestRedis(t *testing.T) {
	site := Site{RedisURL: "redis://localhost:6379/0"}
	if site.redisURL() != "redis://localhost:6379/0" {
		t.Error("wrong REDISURL")
	}

	os.Setenv("REDISURL", "")
	site = Site{RedisURL: ""}
	defer func() {
		recover()
	}()
	site.redisdb()

	t.Error("wrong REDISURL didn't cause any error")
}
