package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Provider struct {
	Alias    string `toml:"-"`
	Type     string `toml:"type"`
	Host     string `toml:"host"`
	Username string `toml:"username"`
	Token    string `toml:"token"`
}

type ProvidersConfig struct {
	Providers map[string]Provider `toml:"providers"`
}

func SneakConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home directory: %w", err)
	}
	dir := filepath.Join(home, ".config", "sneak")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("cannot create config directory: %w", err)
	}
	return dir, nil
}

func providersFilePath() (string, error) {
	dir, err := SneakConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "providers.toml"), nil
}

func LoadProviders() (*ProvidersConfig, error) {
	path, err := providersFilePath()
	if err != nil {
		return nil, err
	}

	cfg := &ProvidersConfig{
		Providers: make(map[string]Provider),
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	// Populate Alias from map key
	normalized := make(map[string]Provider, len(cfg.Providers))
	for alias, p := range cfg.Providers {
		p.Alias = alias
		p.Host = strings.TrimRight(p.Host, "/")
		normalized[alias] = p
	}
	cfg.Providers = normalized

	return cfg, nil
}

func SaveProviders(cfg *ProvidersConfig) error {
	path, err := providersFilePath()
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", path, err)
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	enc.Indent = "\t"
	if err := enc.Encode(cfg); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}

func GetProviderByAlias(alias string) (*Provider, error) {
	cfg, err := LoadProviders()
	if err != nil {
		return nil, err
	}
	p, ok := cfg.Providers[alias]
	if !ok {
		return nil, fmt.Errorf("provider %q not found", alias)
	}
	return &p, nil
}

func GetProviderByHost(host string) (*Provider, error) {
	cfg, err := LoadProviders()
	if err != nil {
		return nil, err
	}
	for _, p := range cfg.Providers {
		if strings.EqualFold(p.Host, host) {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("no provider found for host %q", host)
}
