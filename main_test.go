package traefik_ip2region

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWhistlist(t *testing.T) {
	cfg := CreateConfig()
	cfg.Whitelist.Enabled = true
	// cfg.Whitelist.Country = []string{"中国"}
	cfg.Whitelist.City = []string{"杭州市"}

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "223.5.5.5:9999"
	handler.ServeHTTP(recorder, req)

	if recorder.Result().StatusCode == http.StatusForbidden {
		t.Errorf("invalid status code: %d", recorder.Result().StatusCode)
	}
}

func TestDemo(t *testing.T) {
	cfg := CreateConfig()

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "1.1.1.1:9999"
	handler.ServeHTTP(recorder, req)

	assertHeader(t, req, "X-Ip2region-Country", "澳大利亚")
}

func assertHeader(t *testing.T, req *http.Request, key, expected string) {
	t.Helper()
	k := req.Header.Get(key)
	if k != expected {
		t.Errorf("invalid header value: %s", k)
	}
}
