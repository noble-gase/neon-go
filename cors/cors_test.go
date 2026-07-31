package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newRequest(method, origin string) *http.Request {
	r := httptest.NewRequest(method, "http://example.com/api", nil)
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	return r
}

func serve(c *Cors, r *http.Request) (*httptest.ResponseRecorder, bool) {
	var called bool
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	})
	w := httptest.NewRecorder()
	c.Handler(next).ServeHTTP(w, r)
	return w, called
}

func TestOptionsCloneInputs(t *testing.T) {
	origins := []string{"http://foo.com"}
	methods := []string{http.MethodGet}
	headers := []string{"X-Custom-Header"}
	exposed := []string{"X-Request-ID"}

	c := New(ACAO(origins...), ACAM(methods...), ACAH(headers...), ACEH(exposed...))
	origins[0] = "http://evil.com"
	methods[0] = http.MethodDelete
	headers[0] = "X-Evil-Header"
	exposed[0] = "X-Evil-Response"

	if c.allowOrigins[0] != "http://foo.com" || len(c.allowOrigins) != 1 {
		t.Fatalf("allowOrigins = %v, want cloned explicit origins", c.allowOrigins)
	}
	if c.allowMethods[0] != http.MethodGet || len(c.allowMethods) != 1 {
		t.Fatalf("allowMethods = %v, want cloned explicit methods", c.allowMethods)
	}
	if c.allowHeaders[0] != "X-Custom-Header" || len(c.allowHeaders) != 1 {
		t.Fatalf("allowHeaders = %v, want cloned explicit headers", c.allowHeaders)
	}
	if c.exposeHeaders[0] != "X-Request-ID" || len(c.exposeHeaders) != 1 {
		t.Fatalf("exposeHeaders = %v, want cloned explicit headers", c.exposeHeaders)
	}
}

func TestWildcardOrigin(t *testing.T) {
	w, called := serve(New(), newRequest(http.MethodGet, "http://foo.com"))

	if !called {
		t.Fatal("next handler was not called")
	}
	if got := w.Header().Get(HeaderACAO); got != wildcard {
		t.Fatalf("ACAO = %q, want %q", got, wildcard)
	}
}

func TestAllowedOriginWithCredentials(t *testing.T) {
	c := New(ACAO("http://foo.com"), ACAC(true))
	w, _ := serve(c, newRequest(http.MethodGet, "http://foo.com"))

	if got := w.Header().Get(HeaderACAO); got != "http://foo.com" {
		t.Fatalf("ACAO = %q, want %q", got, "http://foo.com")
	}
	if got := w.Header().Get(HeaderACAC); got != "true" {
		t.Fatalf("ACAC = %q, want %q", got, "true")
	}
}

func TestDisallowedOriginHasNoACAO(t *testing.T) {
	c := New(ACAO("http://foo.com"), ACAC(true))
	w, called := serve(c, newRequest(http.MethodGet, "http://evil.com"))

	if !called {
		t.Fatal("next handler should still run for disallowed origin")
	}
	if got := w.Header().Get(HeaderACAO); got != "" {
		t.Fatalf("ACAO = %q, want empty", got)
	}
}

func TestPreflightRequest(t *testing.T) {
	c := New(ACAO("http://foo.com"))
	r := newRequest(http.MethodOptions, "http://foo.com")
	r.Header.Set("Access-Control-Request-Method", http.MethodPost)

	w, called := serve(c, r)

	if called {
		t.Fatal("next handler must not run for a preflight request")
	}
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestPreflightEchoesRequest(t *testing.T) {
	c := New(ACAO("http://foo.com"))
	r := newRequest(http.MethodOptions, "http://foo.com")
	r.Header.Set("Access-Control-Request-Method", http.MethodPost)
	r.Header.Set("Access-Control-Request-Headers", "X-Custom-Header")

	w, _ := serve(c, r)

	if got := w.Header().Get(HeaderACAM); got != http.MethodPost {
		t.Fatalf("ACAM = %q, want %q", got, http.MethodPost)
	}
	if got := w.Header().Get(HeaderACAH); got != "X-Custom-Header" {
		t.Fatalf("ACAH = %q, want %q", got, "X-Custom-Header")
	}
}

func TestPreflightDisallowedMethod(t *testing.T) {
	c := New(ACAO("http://foo.com"), ACAM(http.MethodGet))
	r := newRequest(http.MethodOptions, "http://foo.com")
	r.Header.Set("Access-Control-Request-Method", http.MethodDelete)

	w, _ := serve(c, r)

	if got := w.Header().Get(HeaderACAO); got != "" {
		t.Fatalf("ACAO = %q, want empty for disallowed method", got)
	}
	if got := w.Header().Get(HeaderACAM); got != "" {
		t.Fatalf("ACAM = %q, want empty for disallowed method", got)
	}
}

func TestPlainOptionsPassesThrough(t *testing.T) {
	c := New(ACAO("http://foo.com"))
	w, called := serve(c, newRequest(http.MethodOptions, "http://foo.com"))

	if !called {
		t.Fatal("plain OPTIONS request must reach the next handler")
	}
	if w.Code == http.StatusNoContent {
		t.Fatal("plain OPTIONS must not be short-circuited with 204")
	}
}

func TestVaryHeaderPreserved(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add(HeaderVary, "Accept-Encoding")
	})
	w := httptest.NewRecorder()
	New().Handler(next).ServeHTTP(w, newRequest(http.MethodGet, "http://foo.com"))

	values := w.Header().Values(HeaderVary)
	if !containsValue(values, "Origin") || !containsValue(values, "Accept-Encoding") {
		t.Fatalf("Vary = %v, want both Origin and Accept-Encoding", values)
	}
}

func containsValue(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}
