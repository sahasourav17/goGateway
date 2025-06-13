package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hashicorp/consul/api"
	"github.com/sahasourav17/goGetway.git/internal/config"
)

func main() {

	consulAddr := os.Getenv("CONSUL_ADDRESS")
	if consulAddr == "" {
		log.Println("CONSUL_ADDRESS not set, defaulting to localhost:8500")
		consulAddr = "localhost:8500"
	}
	consulAddr = strings.Trim(consulAddr, "\"")

	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulAddr

	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		log.Fatalf("Could not create consul client: %v", err)
	}

	kvPair, _, err := consulClient.KV().Get("gateway/config", nil)
	if err != nil {
		log.Fatalf("Failed to fetch config from consul: %v", err)
	}
	if kvPair == nil {
		log.Fatal("Gateway configuration not found in Consul at key 'gateway/config'")
	}

	var cfg config.Config
	if err := json.Unmarshal(kvPair.Value, &cfg); err != nil {
		log.Fatalf("Error parsing config from consul: %v", err)
	}

	gatewayPort := 8080

	// create a chi router
	r := chi.NewRouter()

	// add middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	for _, route := range cfg.Routes {
		service, ok := cfg.Services[route.ServiceName]
		if !ok {
			log.Printf("Service '%s' for route '%s' not found in config, skipping.", route.ServiceName, route.PathPrefix)
			continue
		}

		targetURL, err := url.Parse(service.URL)
		if err != nil {
			log.Printf("Could not parse URL for service '%s': %v", service.Name, err)
			continue
		}

		// reverse proxy for specific service
		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		path := route.PathPrefix
		r.Handle(path+"/*", http.StripPrefix(path, proxy))

		r.Handle(path+"/*", http.StripPrefix(path, proxy))
		log.Printf("Setting up route: %s -> %s", path, service.URL)
	}

	log.Printf("API Gateway listening on port %d...", gatewayPort)

	// start the api gateway server
	if err := http.ListenAndServe(fmt.Sprintf(":%d", gatewayPort), r); err != nil {
		log.Fatalf("Could not start gateway: %v", err)
	}
}
