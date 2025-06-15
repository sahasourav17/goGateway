package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/consul/api"
	"github.com/sahasourav17/goGateway.git/internal/gateway"
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

	redisAddr := os.Getenv("REDIS_ADDRESS")
	if redisAddr == "" {
		log.Println("REDIS_ADDRESS not set, defaulting to localhost:6379")
		redisAddr = "localhost:6379"
	}
	redisAddr = strings.Trim(redisAddr, "\"")

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	gatewayPort := 8080

	gateway.UpdateRouter(consulClient, redisClient)

	go gateway.WatchConsul(consulClient, redisClient)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		gateway.RouterMutex.RLock()
		router := gateway.CurrentRouter
		gateway.RouterMutex.RUnlock()

		if router != nil {
			router.ServeHTTP(w, r)
		} else {
			http.Error(w, "Gateway not yet configured", http.StatusServiceUnavailable)
		}
	})

	log.Printf("API Gateway listening on port %d...", gatewayPort)

	// start the api gateway server
	if err := http.ListenAndServe(fmt.Sprintf(":%d", gatewayPort), nil); err != nil {
		log.Fatalf("Could not start gateway: %v", err)
	}
}
