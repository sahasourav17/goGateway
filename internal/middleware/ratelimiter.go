package middleware

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sahasourav17/goGateway.git/internal/config"
)

func RateLimiter(redisClient *redis.Client, routeConfig *config.RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.Background()

			identifier := r.Header.Get("X-User-ID")
			if identifier == "" {
				identifier = r.RemoteAddr
				ip, _, err := net.SplitHostPort(identifier)
				if err == nil {
					identifier = ip
				}
			}

			routePathPrefix := r.URL.Path

			// Tier based rate limit
			userTier := r.Header.Get("X-User-Tier")
			if userTier == "" {
				userTier = "default"
			}

			tierLimit, ok := routeConfig.Tiers[userTier]
			if !ok {
				tierLimit, ok = routeConfig.Tiers["default"]
				if !ok {
					log.Printf("No rate limit found for route '%s'", routePathPrefix)
					next.ServeHTTP(w, r)
					return

				}
			}

			limit := tierLimit.Requests
			window := time.Duration(tierLimit.Window) * time.Second

			key := fmt.Sprintf("ratelimit:%s:%s", routePathPrefix, identifier)

			now := time.Now().UnixNano()

			windowStart := now - window.Nanoseconds()

			// for handling race conditions if multiple requests arrive at the same time
			pipe := redisClient.TxPipeline()

			pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))

			pipe.ZAdd(ctx, key, &redis.Z{
				Score:  float64(now),
				Member: now,
			})

			countCmd := pipe.ZCard(ctx, key)

			pipe.Expire(ctx, key, window)

			// Execute all commands in the pipeline.
			_, err := pipe.Exec(ctx)

			if err != nil {
				log.Printf("Error executing rate limiter pipeline: %v", err)
				http.Error(w, "Could not process request", http.StatusInternalServerError)
				return
			}

			count := countCmd.Val()
			if count > int64(limit) {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}

			// adding rate limit headers
			remaining := max(int64(limit) - count, 0)
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(window).Unix(), 10))

			// If the limit is not exceeded, call the next handler.
			next.ServeHTTP(w, r)

		})
	}
}
