package router

import (
	"net/http"
)

func New(proxyHandler http.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/", proxyHandler)

	return mux
}
