package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
)

func ServeHTTP(ctx context.Context, port int32) error {
	mux := http.NewServeMux()
	mux.Handle("/echo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))

	mux.Handle("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	}))

	addr := fmt.Sprintf("0.0.0.0:%d", port)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	log.Printf("Listening HTTPServer on %s", addr)

	errChan := make(chan error)
	go func() {
		errChan <- server.ListenAndServe()
	}()

	select {
	case err := <-errChan:
		log.Printf("Failed to serve HTTP: %s, closing", err)
		return nil
	case <-ctx.Done():
		_ = server.Shutdown(context.Background())
		return nil
	}

}
