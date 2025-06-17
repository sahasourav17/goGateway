package config

type MiddlewareConfig struct {
	RateLimit *RateLimitConfig `json:"rate_limit,omitempty"`
}

type RateLimitConfig struct {
	Tiers map[string]TierLimit `json:"tiers"`
}

// TierLimit represents request limit and window for a specific tier
type TierLimit struct {
	Requests int `json:"requests"`
	Window   int `json:"window_seconds"`
}

type Config struct {
	Services map[string]Service `json:"services"`
	Routes   []Route            `json:"routes"`
}

type Service struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Route struct {
	PathPrefix   string            `json:"path_prefix"`
	ServiceName  string            `json:"service_name"`
	AuthRequired bool              `json:"auth_required"`
	Middleware   MiddlewareConfig `json:"middleware,omitempty"`
}
