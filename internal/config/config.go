package config

type Config struct {
	Services map[string]Service `json:"services"`
	Routes   []Route            `json:"routes"`
}

type Service struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Route struct {
	PathPrefix   string `json:"path_prefix"`
	ServiceName  string `json:"service_name"`
	AuthRequired bool   `json:"auth_required"`
}
