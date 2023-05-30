package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

// LoadBalancer represents a basic round-robin load balancer
type LoadBalancer struct {
	backendURLs []*url.URL
	proxy       *httputil.ReverseProxy
	current     int
	mutex       sync.Mutex
}

// NewLoadBalancer creates a new instance of LoadBalancer
func NewLoadBalancer(backendURLs []string) *LoadBalancer {
	urls := make([]*url.URL, len(backendURLs))
	for i, backendURL := range backendURLs {
		u, err := url.Parse(backendURL)
		if err != nil {
			log.Fatal("Failed to parse backend URL:", err)
		}
		urls[i] = u
	}

	return &LoadBalancer{
		backendURLs: urls,
		proxy:       &httputil.ReverseProxy{},
	}
}

// ServeHTTP handles the incoming HTTP requests
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	// Select the next backend server
	lb.current = (lb.current + 1) % len(lb.backendURLs)
	lb.proxy.Director = func(req *http.Request) {
		req.URL.Scheme = lb.backendURLs[lb.current].Scheme
		req.URL.Host = lb.backendURLs[lb.current].Host
		req.URL.Path = lb.backendURLs[lb.current].Path + req.URL.Path
	}

	// Proxy the request to the selected backend server
	lb.proxy.ServeHTTP(w, r)
}

func main() {
	// Configure the load balancer with backend server URLs
	backendURLs := []string{
		"http://localhost:8001",
		"http://localhost:8002",
		"http://localhost:8003",
	}
	lb := NewLoadBalancer(backendURLs)

	// Start the load balancer
	log.Println("Load balancer started on port 8000")
	log.Fatal(http.ListenAndServe(":8000", lb))
}
