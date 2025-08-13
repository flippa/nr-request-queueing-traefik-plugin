package nr_request_queueing_traefik_plugin //nolint:revive,stylecheck

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Config the plugin configuration.
type Config struct {
	// ...
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		// ...
	}
}

// XRequestStart a plugin.
type XRequestStart struct {
	next http.Handler
	name string
	// ...
}

// New created a new plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &XRequestStart{
		next: next,
		name: name,
	}, nil
}

func (e *XRequestStart) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req.Header.Set("X-Request-Start", "t="+unixMilliStr())

	e.next.ServeHTTP(rw, req)
}

func unixMilliStr() string {
	return fmt.Sprintf("%d", time.Now().UnixMilli())
}
