package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

func main() {
	upstream := os.Getenv("UPSTREAM")
	if upstream == "" {
		upstream = "http://localhost:8081"
	}
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	target, err := url.Parse(upstream)
	if err != nil {
		log.Fatalf("bad upstream url: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy error: %s %s: %v", r.Method, r.URL.Path, err)
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}
	proxy.Transport = &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           proxy,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("listening on %s, proxying to %s", addr, upstream)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
