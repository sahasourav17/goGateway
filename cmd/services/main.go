package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {

	port := 8081

	// handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Log the request
		log.Printf("Received request from %s on mock service", r.RemoteAddr)

		// Respond to the client
		fmt.Fprintf(w, "Hello from the Mock Backend Service! You've reached path: %s\n", r.URL.Path)
	})

	log.Printf("Mock service listening on port %d...", port)

	// Start the server
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Could not start mock service: %v", err)
	}
}
