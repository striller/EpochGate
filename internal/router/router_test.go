package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	called := false
	proxy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	h := New(proxy)

	req := httptest.NewRequest("GET", "/some-package", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if !called {
		t.Error("proxy handler was not called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestNew_ForwardsAllPaths(t *testing.T) {
	var receivedPath string
	proxy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	h := New(proxy)

	paths := []string{"/", "/foo", "/foo/bar/baz", "/-/command"}
	for _, path := range paths {
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		if receivedPath != path {
			t.Errorf("path = %q, want %q", receivedPath, path)
		}
	}
}
