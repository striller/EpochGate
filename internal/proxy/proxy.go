package proxy

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/EpochGate/internal/config"
)

type NPMMetadata struct {
	Time map[string]string `json:"time"`
}

type Handler struct {
	proxy    *httputil.ReverseProxy
	target   *url.URL
	registry string
	minAge   float64
	cache    sync.Map
}

func New(cfg *config.Config) (*Handler, error) {
	target, err := url.Parse(cfg.NexusURL)
	if err != nil {
		return nil, err
	}

	return &Handler{
		proxy:    httputil.NewSingleHostReverseProxy(target),
		target:   target,
		registry: cfg.NPMRegistry,
		minAge:   cfg.MinAgeDays,
	}, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	packageName := parts[0]

	if strings.HasPrefix(packageName, "-") {
		h.proxy.ServeHTTP(w, r)
		return
	}

	if result, ok := h.cache.Load(packageName); ok {
		if isAllowed := result.(bool); !isAllowed {
			h.sendBlockResponse(w, packageName, h.minAge)
			return
		}
		h.proxy.ServeHTTP(w, r)
		return
	}

	resp, err := http.Get(h.registry + packageName)
	if err != nil || resp.StatusCode != http.StatusOK {
		h.proxy.ServeHTTP(w, r)
		return
	}
	defer resp.Body.Close()

	var metadata NPMMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		h.proxy.ServeHTTP(w, r)
		return
	}

	lastModifiedStr, exists := metadata.Time["modified"]
	if exists {
		lastModifiedTime, err := time.Parse(time.RFC3339Nano, lastModifiedStr)
		if err == nil {
			ageInDays := time.Since(lastModifiedTime).Hours() / 24

			if ageInDays < h.minAge {
				slog.Info("blocked package", "package", packageName, "age_days", ageInDays)
				h.cache.Store(packageName, false)
				h.sendBlockResponse(w, packageName, ageInDays)
				return
			}
		}
	}

	slog.Info("allowed package", "package", packageName)
	h.cache.Store(packageName, true)

	r.Host = h.target.Host
	h.proxy.ServeHTTP(w, r)
}

func (h *Handler) sendBlockResponse(w http.ResponseWriter, pkg string, age float64) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	msg := fmt.Sprintf(`{"error": "Package '%s' is under quarantine (only %.1f days old)."}`, pkg, age)
	w.Write([]byte(msg))
}


