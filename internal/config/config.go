package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Token     string
	ScopeType string
	ScopeID   string
	APIURL    string
	Debug     bool
}

type sanityConfig struct {
	AuthToken string `json:"authToken"`
}

func Load(flagToken, flagProject, flagAPIURL string, staging bool) (Config, error) {
	cfg := Config{
		ScopeType: "project",
		APIURL:    "https://api.sanity.io",
	}
	if staging {
		cfg.APIURL = "https://api.sanity.work"
	}

	cfg.Token = resolve(flagToken, "SANITY_AUTH_TOKEN", "")
	if cfg.Token == "" {
		t, err := readSanityToken(staging)
		if err == nil {
			cfg.Token = t
		}
	}
	if cfg.Token == "" {
		return cfg, fmt.Errorf("no auth token found (use --token, SANITY_AUTH_TOKEN, or log in with the Sanity CLI)")
	}

	cfg.ScopeID = resolve(flagProject, "SANITY_PROJECT_ID", "")
	if cfg.ScopeID == "" {
		return cfg, fmt.Errorf("no project ID provided (use --project or SANITY_PROJECT_ID)")
	}

	if u := resolve(flagAPIURL, "BLUEPRINTS_API_URL", ""); u != "" {
		cfg.APIURL = u
	}

	return cfg, nil
}

func resolve(flag, envKey, fallback string) string {
	if flag != "" {
		return flag
	}
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return fallback
}

func readSanityToken(staging bool) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := "sanity"
	if staging {
		configDir = "sanity-staging"
	}
	data, err := os.ReadFile(filepath.Join(home, ".config", configDir, "config.json"))
	if err != nil {
		return "", err
	}
	var sc sanityConfig
	if err := json.Unmarshal(data, &sc); err != nil {
		return "", err
	}
	return sc.AuthToken, nil
}
