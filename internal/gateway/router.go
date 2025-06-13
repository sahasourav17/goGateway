package gateway

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hashicorp/consul/api"
	"github.com/sahasourav17/goGetway.git/internal/config"
)

var (
	// CurrentRouter is the currently active router that serves traffic
	CurrentRouter http.Handler
	// RouterMutex protects CurrentRouter during hot reloading
	RouterMutex sync.RWMutex
)

const consulKey = "gateway/config"

func UpdateRouter(consulClient *api.Client) {
	log.Println("Updating router configuration from Consul...")
	kvPair, _, err := consulClient.KV().Get(consulKey, nil)
	if err != nil || kvPair == nil {
		log.Printf("Could not fetch or find config in Consul ('%s'): %v", consulKey, err)
		return
	}

	var cfg config.Config
	if err := json.Unmarshal(kvPair.Value, &cfg); err != nil {
		log.Printf("Error parsing config from consul: %v", err)
		return
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	for _, route := range cfg.Routes {
		service, ok := cfg.Services[route.ServiceName]
		if !ok {
			log.Printf("Service '%s' for route '%s' not found, skipping.", route.ServiceName, route.PathPrefix)
			continue
		}

		targetURL, err := url.Parse(service.URL)
		if err != nil {
			log.Printf("Invalid URL for service '%s': %v", service.Name, err)
			continue
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		path := route.PathPrefix
		log.Println("Adding route:", path)
		r.Handle(path+"/*", http.StripPrefix(path, proxy))
	}

	// Replace the current router with the new one
	RouterMutex.Lock()
	CurrentRouter = r
	RouterMutex.Unlock()

	log.Println("Router configuration updated successfully.")
}

func WatchConsul(consulClient *api.Client) {
	var lastIndex uint64
	for {
		opts := &api.QueryOptions{
			WaitIndex: lastIndex,
		}
		kvPair, meta, err := consulClient.KV().Get(consulKey, opts)
		if err != nil {
			log.Printf("Error watching consul: %v", err)
			time.Sleep(5 * time.Second) // Wait before retrying
			continue
		}

		// if index is different, it means the config has changed
		if meta.LastIndex != lastIndex {
			if kvPair != nil {
				UpdateRouter(consulClient)
				lastIndex = meta.LastIndex
			}
		}
	}
}
