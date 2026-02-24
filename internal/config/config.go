package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Log      LogConfig      `yaml:"log"`
}

type ServerConfig struct {
	Port            int           `yaml:"port"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type DatabaseConfig struct {
	URL             string        `yaml:"url"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// Load загружает конфигурацию из файла и переменных окружения
func Load(path string) (*Config, error) {
	cfg := &Config{}

	// Загружаем из файла если указан
	if path != "" {
		if err := cfg.loadFromFile(path); err != nil {
			return nil, fmt.Errorf("failed to load config from file: %w", err)
		}
	}

	// Переопределяем из ENV
	cfg.loadFromEnv()

	// Устанавливаем значения по умолчанию
	cfg.setDefaults()

	// Валидируем
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) loadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	expanded := os.ExpandEnv(string(data))

	return yaml.Unmarshal([]byte(expanded), c)
}

func (c *Config) loadFromEnv() {
	// Server
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Server.Port = p
		}
	}
	if timeout := os.Getenv("SERVER_READ_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			c.Server.ReadTimeout = d
		}
	}
	if timeout := os.Getenv("SERVER_WRITE_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			c.Server.WriteTimeout = d
		}
	}
	if timeout := os.Getenv("SERVER_SUTDOWN_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			c.Server.ShutdownTimeout = d
		}
	}

	// Database
	if url := os.Getenv("DATABASE_URL"); url != "" {
		c.Database.URL = url
	}

	if oConns := os.Getenv("DB_MAX_OPEN_CONNS"); oConns != "" {
		if n, err := strconv.Atoi(oConns); err == nil {
			c.Database.MaxOpenConns = n
		}
	}
	if iConns := os.Getenv("DB_MAX_IDLE_CONNS"); iConns != "" {
		if n, err := strconv.Atoi(iConns); err == nil {
			c.Database.MaxIdleConns = n
		}
	}
	if lifetime := os.Getenv("DB_CONNS_MAX_LIFETIME"); lifetime != "" {
		if d, err := time.ParseDuration(lifetime); err == nil {
			c.Database.ConnMaxLifetime = d
		}
	}

	// Log
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		c.Log.Level = level
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		c.Log.Format = format
	}
}

func (c *Config) setDefaults() {
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 10 * time.Second
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 10 * time.Second
	}
	if c.Server.ShutdownTimeout == 0 {
		c.Server.ShutdownTimeout = 10 * time.Second
	}

	if c.Database.URL == "" {
		c.Database.URL = "postgres://postgres:postgres@localhost:5432/<base>?sslmode=disable"
	}
	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = 25
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = 5
	}
	if c.Database.ConnMaxLifetime == 0 {
		c.Database.ConnMaxLifetime = 5 * time.Minute
	}

	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.Log.Format == "" {
		c.Log.Format = "json"
	}
}

func (c *Config) validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Database.URL == "" {
		return fmt.Errorf("database URL is required")
	}

	return nil
}
