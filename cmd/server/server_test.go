package main

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestServer_ListenAndServe_GracefulShutdown(t *testing.T) {
	srv := &server{
		addr:    ":0",
		handler: http.NewServeMux(),
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- srv.ListenAndServe(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("ListenAndServe() error = %v, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("ListenAndServe() did not return in time")
	}
}

func TestServer_ListenAndServe_AddressInUse(t *testing.T) {
	blocker := &http.Server{
		Addr:    ":19876",
		Handler: http.NewServeMux(),
	}
	go blocker.ListenAndServe()
	defer blocker.Close()

	time.Sleep(50 * time.Millisecond)

	srv := &server{
		addr:    ":19876",
		handler: http.NewServeMux(),
	}

	ctx := context.Background()
	err := srv.ListenAndServe(ctx)
	if err == nil {
		t.Error("ListenAndServe() expected error for address in use")
	}
}
