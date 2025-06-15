package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: go run main.go <service-name> <port>")
	}
	serviceName := os.Args[1]
	port := os.Args[2]

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request received on %s", serviceName)
		userID := r.Header.Get("X-User-ID")

		// If the header is present and not empty, it means the request was authenticated.
		if userID != "" {
			fmt.Fprintf(w, "Hello from the %s! You are authenticated as user: %s. Path: %s\n", serviceName, userID, r.URL.Path)
		} else {
			// If the header is missing, it was a public request.
			fmt.Fprintf(w, "Hello from the %s! This is a public endpoint. Path: %s\n", serviceName, r.URL.Path)
		}
	})

	log.Printf("%s listening on port %s...", serviceName, port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Could not start mock service: %v", err)
	}
}
