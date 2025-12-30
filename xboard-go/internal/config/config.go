package config

import (
	"encoding/json"
	"os"
	"time"
)

type Config struct {
	Server     ServerConfig     `json:"server"`
	Database   DatabaseConfig   `json:"database"`
	Auth       AuthConfig       `json:"auth"`
	Node       NodeConfig       `json:"node"`
	Prometheus PrometheusConfig `json:"prometheus"`
	Telegram   TelegramConfig   `json:"telegram"`
}

type ServerConfig struct {
	Port string `json:"port"`
	Mode string `json:"mode"`
}

type DatabaseConfig struct {
	Host         string `json:"host"`
	Port         string `json:"port"`
	User         string `json:"user"`
	Password     string `json:"password"`
	DBName       string `json:"dbname"`
	MaxIdleConns int    `json:"max_idle_conns"`
	MaxOpenConns int    `json:"max_open_conns"`
}

type AuthConfig struct {
	JWTSecret            string `json:"jwt_secret"`
	AccessTokenDuration  string `json:"access_token_duration"`
	RefreshTokenDuration string `json:"refresh_token_duration"`
}

func (a *AuthConfig) GetAccessTokenDuration() time.Duration {
	d, err := time.ParseDuration(a.AccessTokenDuration)
	if err != nil {
		return 15 * time.Minute
	}
	return d
}

func (a *AuthConfig) GetRefreshTokenDuration() time.Duration {
	d, err := time.ParseDuration(a.RefreshTokenDuration)
	if err != nil {
		return 168 * time.Hour
	}
	return d
}

type NodeConfig struct {
	ServerToken  string `json:"server_token"`
	PullInterval int    `json:"pull_interval"`
	PushInterval int    `json:"push_interval"`
}

type PrometheusConfig struct {
	URL string `json:"url"`
}

type TelegramConfig struct {
	Token          string `json:"token"`
	PollingTimeout int    `json:"polling_timeout"`
}

func Load(configPath string) (*Config, error) {
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}

	// Override with environment variables if present
	if port := os.Getenv("SERVER_PORT"); port != "" {
		cfg.Server.Port = port
	}
	if mode := os.Getenv("SERVER_MODE"); mode != "" {
		cfg.Server.Mode = mode
	}
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		cfg.Database.Port = port
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		cfg.Database.Password = password
	}
	if dbname := os.Getenv("DB_NAME"); dbname != "" {
		cfg.Database.DBName = dbname
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.Auth.JWTSecret = secret
	}
	if token := os.Getenv("NODE_SERVER_TOKEN"); token != "" {
		cfg.Node.ServerToken = token
	}
	if promURL := os.Getenv("PROMETHEUS_URL"); promURL != "" {
		cfg.Prometheus.URL = promURL
	}
	if tgToken := os.Getenv("TELEGRAM_TOKEN"); tgToken != "" {
		cfg.Telegram.Token = tgToken
	}

	return &cfg, nil
}
