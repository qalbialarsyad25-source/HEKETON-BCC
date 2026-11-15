package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

const DefaultCORSOrigins = "http://localhost:3000,http://localhost:5173"

type HTTPConfig struct {
	CORSAllowedOrigins []string
	CookieDomain       string
	CookieSecure       bool
	TrustedProxies     []string
}

func LoadHTTPConfig(app *AppConfig) (*HTTPConfig, error) {
	var errs []error

	rawOrigins := GetEnv("CORS_ALLOWED_ORIGINS", "")
	if rawOrigins == "" && !app.IsProduction() {
		rawOrigins = DefaultCORSOrigins
	}
	origins := splitList(rawOrigins)
	if len(origins) == 0 && app.IsProduction() {
		errs = append(errs, errors.New("CORS_ALLOWED_ORIGINS: required in production (comma-separated frontend origins)"))
	}
	for _, origin := range origins {
		errs = append(errs, checkOrigin(origin))
	}

	cookieDomain := GetEnv("COOKIE_DOMAIN", "")
	cookieSecure, err := GetEnvBool("COOKIE_SECURE", app.IsProduction())
	errs = append(errs, err)

	trustedProxies := splitList(GetEnv("TRUSTED_PROXIES", ""))
	for _, proxy := range trustedProxies {
		errs = append(errs, checkProxy(proxy))
	}

	return &HTTPConfig{
		CORSAllowedOrigins: origins,
		CookieDomain:       cookieDomain,
		CookieSecure:       cookieSecure,
		TrustedProxies:     trustedProxies,
	}, errors.Join(errs...)
}

func checkOrigin(origin string) error {
	if strings.Contains(origin, "*") {
		return fmt.Errorf("CORS_ALLOWED_ORIGINS: %q: a wildcard is not allowed (credentials are on), list each origin", origin)
	}

	u, err := url.Parse(origin)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" ||
		u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.User != nil || strings.HasSuffix(origin, "?") {
		return fmt.Errorf("CORS_ALLOWED_ORIGINS: %q must look like https://app.example.com or http://localhost:3000 (scheme and host, no path or trailing slash)", origin)
	}

	return nil
}
func checkProxy(proxy string) error {
	if net.ParseIP(proxy) != nil {
		return nil
	}
	if _, _, err := net.ParseCIDR(proxy); err == nil {
		return nil
	}
	return fmt.Errorf("TRUSTED_PROXIES: %q must be an IP or a CIDR (e.g. 10.0.0.1 or 10.0.0.0/8)", proxy)
}

func splitList(value string) []string {
	var items []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}
