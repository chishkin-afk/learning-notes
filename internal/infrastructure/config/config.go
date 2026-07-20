package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
)

var (
	ErrInvalidServerCert = errors.New("invalid path to server cert")
	ErrInvalidServerKey  = errors.New("invalid server key")
)

// Config is a global settings for project
//
// This structure contains all settings for all external and internal components,
// such as a server or a database.
type Config struct {
	App         App         `yaml:"app"`
	Server      Server      `yaml:"server"`
	Persistence Persistence `yaml:"persistence"`
	Cache       Cache       `yaml:"cache"`
}

type App struct {
	Name    string `yaml:"name" validate:"required"`
	Version string `yaml:"version" validate:"required,semver"`
	Env     string `yaml:"env" validate:"required,oneof=dev prod local"`
}

type Server struct {
	HTTP struct {
		Addr string `yaml:"addr" validate:"required,hostname_port"`

		// TLS is a struct field of server for tls settings
		//
		// This field has no validation because it follows its own specific logic
		// validation takes place in a separate function.
		TLS struct {
			Enable         bool   `yaml:"enable"`
			ServerCertPath string `yaml:"server_cert_path"`
			ServerKeyPath  string `yaml:"server_key_path"`
		} `yaml:"tls"`
	} `yaml:"http"`
	Conns struct {
		ReadTimeout  time.Duration `yaml:"read_timeout" validate:"required,min=100ms"`
		WriteTimeout time.Duration `yaml:"write_timeout" validate:"required,min=100ms"`
		IdleTimeout  time.Duration `yaml:"idle_timeout" validate:"required,min=100ms"`
	} `yaml:"conns"`
}

type Persistence struct {
	Postgres struct {
		Host    string `yaml:"host" validate:"required,hostname"`
		Port    int    `yaml:"port" validate:"required,gte=1,lte=65535"`
		SSLMode string `yaml:"sslmode" validate:"required,oneof=disable enable"`
		Auth    struct {
			User     string `yaml:"user" validate:"required"`
			Password string `yaml:"password" validate:"required"`
			DB       string `yaml:"db" validate:"required"`
		} `yaml:"auth"`
		Conns struct {
			MaxIdles    int           `yaml:"max_idles" validate:"required,gte=1"`
			MaxOpens    int           `yaml:"max_opens" validate:"required,gte=1"`
			MaxIdleTime time.Duration `yaml:"max_idle_time" validate:"required,min=100ms"`
			MaxLifetime time.Duration `yaml:"max_lifetime" validate:"required,min=100ms"`
		} `yaml:"conns"`
	} `yaml:"postgres"`
}

type Cache struct {
	NoteTTL time.Duration `yaml:"note_ttl" validate:"required,min=1m"`
	Redis   struct {
		Host string `yaml:"host" validate:"required,hostname"`
		Port int    `yaml:"port" validate:"required,gte=1,lte=65535"`

		// Auth is a special struct field of cache
		//
		// There is no validation here because these parameters in Redis
		// can be left empty for the default user.
		Auth struct {
			Username string `yaml:"username"`
			Password string `yaml:"password"`
			DB       int    `yaml:"db"`
		}
	} `yaml:"redis"`
}

// New returns the *Config settings
//
// It takes the path to a config file as input, parses it,
// and returns either an error or a pointer to the config.
// It assumes that the .env file has already been loaded.
func New(path string) (*Config, error) {
	bytes, err := loadBytes(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg, err := parseBytes(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return cfg, nil
}

func validate(cfg *Config) error {
	validate := validator.New()

	if err := validate.Struct(cfg); err != nil {
		return err
	}

	return validateTLS(&cfg.Server)
}

func validateTLS(server *Server) error {
	if !server.HTTP.TLS.Enable {
		return nil
	}

	serverCertPath := filepath.Clean(server.HTTP.TLS.ServerCertPath)
	if stat, err := os.Stat(serverCertPath); err != nil || stat.IsDir() {
		return fmt.Errorf("%w: %s", ErrInvalidServerCert, serverCertPath)
	}

	serverKeyPath := filepath.Clean(server.HTTP.TLS.ServerKeyPath)
	if stat, err := os.Stat(serverKeyPath); err != nil || stat.IsDir() {
		return fmt.Errorf("%w: %s", ErrInvalidServerKey, serverKeyPath)
	}

	return nil
}

func parseBytes(bytes []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func loadBytes(path string) ([]byte, error) {
	path = filepath.Clean(path)
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := os.ExpandEnv(string(bytes))
	return []byte(content), nil
}
