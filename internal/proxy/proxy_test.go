package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/EpochGate/internal/config"
)

func newTestHandler(t *testing.T, registryURL, nexusURL string, minAge float64) *Handler {
	t.Helper()
	cfg := &config.Config{
		NexusURL:    nexusURL,
		NPMRegistry: registryURL,
		MinAgeDays:  minAge,
		ListenPort:  ":0",
	}
	h, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return h
}

func TestNew_InvalidURL(t *testing.T) {
	cfg := &config.Config{NexusURL: "://invalid"}
	_, err := New(cfg)
	if err == nil {
		t.Error("New() expected error for invalid URL")
	}
}

func TestServeHTTP_NPMCommand(t *testing.T) {
	nexus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("proxied"))
	}))
	defer nexus.Close()

	h := newTestHandler(t, "http://unused/", nexus.URL, 7)

	req := httptest.NewRequest("GET", "/-/package", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestServeHTTP_CachedBlocked(t *testing.T) {
	nexus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach nexus for blocked package")
	}))
	defer nexus.Close()

	h := newTestHandler(t, "http://unused/", nexus.URL, 7)
	h.cache.Store("blocked-pkg", false)

	req := httptest.NewRequest("GET", "/blocked-pkg", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestServeHTTP_CachedAllowed(t *testing.T) {
	nexus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("proxied"))
	}))
	defer nexus.Close()

	h := newTestHandler(t, "http://unused/", nexus.URL, 7)
	h.cache.Store("allowed-pkg", true)

	req := httptest.NewRequest("GET", "/allowed-pkg", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestServeHTTP_RegistryError(t *testing.T) {
	nexus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("proxied"))
	}))
	defer nexus.Close()

	h := newTestHandler(t, "http://localhost:1/", nexus.URL, 7)

	req := httptest.NewRequest("GET", "/some-pkg", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestServeHTTP_RegistryNon200(t *testing.T) {
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer registry.Close()

	nexus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer nexus.Close()

	h := newTestHandler(t, registry.URL+"/", nexus.URL, 7)

	req := httptest.NewRequest("GET", "/missing-pkg", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestServeHTTP_InvalidJSON(t *testing.T) {
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json"))
	}))
	defer registry.Close()

	nexus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer nexus.Close()

	h := newTestHandler(t, registry.URL+"/", nexus.URL, 7)

	req := httptest.NewRequest("GET", "/bad-json", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestServeHTTP_MissingModifiedField(t *testing.T) {
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(NPMMetadata{
			Time: map[string]string{"created": "2020-01-01T00:00:00Z"},
		})
	}))
	defer registry.Close()

	nexus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer nexus.Close()

	h := newTestHandler(t, registry.URL+"/", nexus.URL, 7)

	req := httptest.NewRequest("GET", "/no-modified", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestServeHTTP_InvalidDateFormat(t *testing.T) {
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(NPMMetadata{
			Time: map[string]string{"modified": "not-a-date"},
		})
	}))
	defer registry.Close()

	nexus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer nexus.Close()

	h := newTestHandler(t, registry.URL+"/", nexus.URL, 7)

	req := httptest.NewRequest("GET", "/bad-date", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestServeHTTP_BlockedYoungPackage(t *testing.T) {
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(NPMMetadata{
			Time: map[string]string{"modified": time.Now().Add(-2 * 24 * time.Hour).Format(time.RFC3339Nano)},
		})
	}))
	defer registry.Close()

	nexus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach nexus for blocked package")
	}))
	defer nexus.Close()

	h := newTestHandler(t, registry.URL+"/", nexus.URL, 7)

	req := httptest.NewRequest("GET", "/young-pkg", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestServeHTTP_AllowedOldPackage(t *testing.T) {
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(NPMMetadata{
			Time: map[string]string{"modified": time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339Nano)},
		})
	}))
	defer registry.Close()

	nexus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("proxied"))
	}))
	defer nexus.Close()

	h := newTestHandler(t, registry.URL+"/", nexus.URL, 7)

	req := httptest.NewRequest("GET", "/old-pkg", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestServeHTTP_BlockResponseJSON(t *testing.T) {
	nexus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach nexus")
	}))
	defer nexus.Close()

	h := newTestHandler(t, "http://unused/", nexus.URL, 7)
	h.cache.Store("blocked-pkg", false)

	req := httptest.NewRequest("GET", "/blocked-pkg", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", w.Header().Get("Content-Type"))
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Error("response body missing 'error' key")
	}
}

func TestServeHTTP_BlockedPackageCachedAsBlocked(t *testing.T) {
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(NPMMetadata{
			Time: map[string]string{"modified": time.Now().Add(-1 * 24 * time.Hour).Format(time.RFC3339Nano)},
		})
	}))
	defer registry.Close()

	nexus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer nexus.Close()

	h := newTestHandler(t, registry.URL+"/", nexus.URL, 7)

	req := httptest.NewRequest("GET", "/new-pkg", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("first request status = %d, want %d", w.Code, http.StatusForbidden)
	}

	val, ok := h.cache.Load("new-pkg")
	if !ok {
		t.Fatal("package not cached")
	}
	if val.(bool) != false {
		t.Error("package should be cached as blocked")
	}
}

func TestServeHTTP_AllowedPackageCachedAsAllowed(t *testing.T) {
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(NPMMetadata{
			Time: map[string]string{"modified": time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339Nano)},
		})
	}))
	defer registry.Close()

	nexus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer nexus.Close()

	h := newTestHandler(t, registry.URL+"/", nexus.URL, 7)

	req := httptest.NewRequest("GET", "/trusted-pkg", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	val, ok := h.cache.Load("trusted-pkg")
	if !ok {
		t.Fatal("package not cached")
	}
	if val.(bool) != true {
		t.Error("package should be cached as allowed")
	}
}

func TestServeHTTP_PackageExactlyAtMinAge(t *testing.T) {
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(NPMMetadata{
			Time: map[string]string{"modified": time.Now().Add(-7 * 24 * time.Hour).Format(time.RFC3339Nano)},
		})
	}))
	defer registry.Close()

	nexus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer nexus.Close()

	h := newTestHandler(t, registry.URL+"/", nexus.URL, 7)

	req := httptest.NewRequest("GET", "/exact-age-pkg", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}
