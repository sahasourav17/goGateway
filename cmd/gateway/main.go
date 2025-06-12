package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	gatewayPort := 8080

	targetServiceUrl, err := url.Parse("http://localhost:8081")
	if err != nil {
		log.Fatalf("Invalid target service URL: %v", err)
	}

	// create a reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetServiceUrl)

	// create a chi router
	r := chi.NewRouter()

	// add middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	// define route to proxy
	routeToProxy := "/api/users"

	r.Handle(routeToProxy+"/*", http.StripPrefix(routeToProxy, proxy))

	log.Printf("APIGateway listening on port %d...", gatewayPort)

	// start the api gateway server
	if err := http.ListenAndServe(fmt.Sprintf(":%d", gatewayPort), r); err != nil {
		log.Fatalf("Could not start gateway: %v", err)
	}
}
