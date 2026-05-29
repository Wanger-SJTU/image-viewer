package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Port              int
	DBPath            string
	CacheDir          string
	SupportedRawExts  []string
	SupportedJpgExts  []string
	ConcurrencyLimit  int
}

func Load() *Config {
	cfg := &Config{
		Port:             8080,
		DBPath:           "storage/viewer.db",
		CacheDir:         "storage/cache",
		ConcurrencyLimit: 4,
		SupportedRawExts: []string{".CR3", ".ARW", ".NEF", ".CR2", ".DNG"},
		SupportedJpgExts: []string{".JPG", ".JPEG", ".jpg", ".jpeg"},
	}

	if port := os.Getenv("VIEWER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil && p > 0 {
			cfg.Port = p
		}
	}

	if dbPath := os.Getenv("VIEWER_DB_PATH"); dbPath != "" {
		cfg.DBPath = dbPath
	}

	if cacheDir := os.Getenv("VIEWER_CACHE_DIR"); cacheDir != "" {
		cfg.CacheDir = cacheDir
	}

	if limit := os.Getenv("VIEWER_CONCURRENCY"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			cfg.ConcurrencyLimit = l
		}
	}

	if rawExts := os.Getenv("VIEWER_RAW_EXTS"); rawExts != "" {
		cfg.SupportedRawExts = splitAndNormalize(rawExts)
	}

	if jpgExts := os.Getenv("VIEWER_JPG_EXTS"); jpgExts != "" {
		cfg.SupportedJpgExts = splitAndNormalize(jpgExts)
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(cfg.CacheDir, 0755); err != nil {
		panic("failed to create cache directory: " + err.Error())
	}

	// Ensure DB parent directory exists
	dbDir := filepath.Dir(cfg.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		panic("failed to create database directory: " + err.Error())
	}

	return cfg
}

func splitAndNormalize(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			if !strings.HasPrefix(p, ".") {
				p = "." + p
			}
			result = append(result, strings.ToLower(p))
		}
	}
	return result
}
