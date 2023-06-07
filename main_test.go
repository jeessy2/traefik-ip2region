package traefik_ip2region

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
