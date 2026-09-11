// Package config loads and validates the MCP WordPress connection settings.
package config

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimeout        = 30 * time.Second
	defaultMaxUploadBytes = int64(25 << 20) // 25 MiB
)

// Config is the runtime configuration for the WordPress REST API client.
type Config struct {
	BaseURL        *url.URL
	Username       string
	AppPassword    string
	Timeout        time.Duration
	MaxUploadBytes int64
	UploadRoot     string
}

// Load reads configuration from environment variables. It never logs credentials.
func Load() (Config, error) {
	baseURL, err := parseBaseURL(os.Getenv("WP_BASE_URL"))
	if err != nil {
		return Config{}, err
	}

	username := strings.TrimSpace(os.Getenv("WP_USERNAME"))
	if username == "" {
		return Config{}, errors.New("WP_USERNAME is required")
	}
	password := strings.TrimSpace(os.Getenv("WP_APP_PASSWORD"))
	if password == "" {
		return Config{}, errors.New("WP_APP_PASSWORD is required")
	}

	timeout := defaultTimeout
	if raw := strings.TrimSpace(os.Getenv("WP_TIMEOUT")); raw != "" {
		timeout, err = time.ParseDuration(raw)
		if err != nil || timeout <= 0 {
			return Config{}, fmt.Errorf("WP_TIMEOUT must be a positive duration: %w", err)
		}
	}

	maxUploadBytes := defaultMaxUploadBytes
	if raw := strings.TrimSpace(os.Getenv("WP_MAX_UPLOAD_BYTES")); raw != "" {
		maxUploadBytes, err = strconv.ParseInt(raw, 10, 64)
		if err != nil || maxUploadBytes <= 0 {
			return Config{}, fmt.Errorf("WP_MAX_UPLOAD_BYTES must be a positive integer: %w", err)
		}
	}

	var uploadRoot string
	if raw := strings.TrimSpace(os.Getenv("WP_UPLOAD_ROOT")); raw != "" {
		uploadRoot, err = filepath.EvalSymlinks(raw)
		if err != nil {
			return Config{}, fmt.Errorf("resolve WP_UPLOAD_ROOT: %w", err)
		}
		uploadRoot, err = filepath.Abs(uploadRoot)
		if err != nil {
			return Config{}, fmt.Errorf("make WP_UPLOAD_ROOT absolute: %w", err)
		}
		info, err := os.Stat(uploadRoot)
		if err != nil {
			return Config{}, fmt.Errorf("stat WP_UPLOAD_ROOT: %w", err)
		}
		if !info.IsDir() {
			return Config{}, errors.New("WP_UPLOAD_ROOT must be a directory")
		}
	}

	return Config{
		BaseURL:        baseURL,
		Username:       username,
		AppPassword:    password,
		Timeout:        timeout,
		MaxUploadBytes: maxUploadBytes,
		UploadRoot:     uploadRoot,
	}, nil
}

func parseBaseURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("WP_BASE_URL is required")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, errors.New("WP_BASE_URL must be an absolute URL")
	}
	if u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("WP_BASE_URL must not contain credentials, a query, or a fragment")
	}
	if u.Scheme != "https" && !(u.Scheme == "http" && isLoopbackHost(u.Hostname())) {
		return nil, errors.New("WP_BASE_URL must use HTTPS (HTTP is allowed only for localhost)")
	}
	u.Path = strings.TrimRight(u.Path, "/")
	return u, nil
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// LoadEnvFile adds simple KEY=VALUE entries to the process environment, without
// overriding values already supplied by the launcher. It deliberately does not
// implement shell expansion or command substitution.
func LoadEnvFile(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("%s:%d: expected KEY=VALUE", name, lineNo)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || strings.ContainsAny(key, " \t") {
			return fmt.Errorf("%s:%d: invalid variable name", name, lineNo)
		}
		if len(value) >= 2 && ((value[0] == '\'' && value[len(value)-1] == '\'') || (value[0] == '"' && value[len(value)-1] == '"')) {
			value = value[1 : len(value)-1]
		}
		if _, alreadySet := os.LookupEnv(key); !alreadySet {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("%s:%d: set environment: %w", name, lineNo, err)
			}
		}
	}
	return scanner.Err()
}
