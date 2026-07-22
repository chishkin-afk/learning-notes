package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBytes(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		data := []byte(`
app:
  name: notes
  version: 1.0.0
  env: dev
`)

		cfg, err := parseBytes(data)

		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, "notes", cfg.App.Name)
		assert.Equal(t, "1.0.0", cfg.App.Version)
		assert.Equal(t, "dev", cfg.App.Env)
	})

	t.Run("invalid yaml", func(t *testing.T) {
		_, err := parseBytes([]byte("app: ["))

		require.Error(t, err)
	})
}

func TestLoadBytes(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir := t.TempDir()

		path := filepath.Join(dir, "config.yaml")

		require.NoError(t, os.Setenv("APP_NAME", "notes"))
		t.Cleanup(func() {
			os.Unsetenv("APP_NAME")
		})

		require.NoError(
			t,
			os.WriteFile(
				path,
				[]byte("name: ${APP_NAME}"),
				0644,
			),
		)

		bytes, err := loadBytes(path)

		require.NoError(t, err)
		assert.Equal(t, "name: notes", string(bytes))
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := loadBytes("missing.yaml")

		require.Error(t, err)
	})
}

func TestValidateTLS(t *testing.T) {
	t.Run("tls disabled", func(t *testing.T) {
		var server Server

		err := validateTLS(&server)

		require.NoError(t, err)
	})

	t.Run("valid cert and key", func(t *testing.T) {
		dir := t.TempDir()

		cert := filepath.Join(dir, "cert.pem")
		key := filepath.Join(dir, "key.pem")

		require.NoError(t, os.WriteFile(cert, []byte("cert"), 0644))
		require.NoError(t, os.WriteFile(key, []byte("key"), 0644))

		var server Server
		server.HTTP.TLS.Enable = true
		server.HTTP.TLS.ServerCertPath = cert
		server.HTTP.TLS.ServerKeyPath = key

		err := validateTLS(&server)

		require.NoError(t, err)
	})

	t.Run("invalid cert", func(t *testing.T) {
		dir := t.TempDir()

		key := filepath.Join(dir, "key.pem")
		require.NoError(t, os.WriteFile(key, []byte("key"), 0644))

		var server Server
		server.HTTP.TLS.Enable = true
		server.HTTP.TLS.ServerCertPath = filepath.Join(dir, "missing.pem")
		server.HTTP.TLS.ServerKeyPath = key

		err := validateTLS(&server)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidServerCert)
	})

	t.Run("invalid key", func(t *testing.T) {
		dir := t.TempDir()

		cert := filepath.Join(dir, "cert.pem")
		require.NoError(t, os.WriteFile(cert, []byte("cert"), 0644))

		var server Server
		server.HTTP.TLS.Enable = true
		server.HTTP.TLS.ServerCertPath = cert
		server.HTTP.TLS.ServerKeyPath = filepath.Join(dir, "missing.pem")

		err := validateTLS(&server)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidServerKey)
	})
}

func TestNew(t *testing.T) {
	dir := t.TempDir()

	cert := filepath.Join(dir, "cert.pem")
	key := filepath.Join(dir, "key.pem")

	require.NoError(t, os.WriteFile(cert, []byte("cert"), 0644))
	require.NoError(t, os.WriteFile(key, []byte("key"), 0644))

	cfgPath := filepath.Join(dir, "config.yaml")

	cfg := `
app:
  name: notes
  version: 1.0.0
  env: dev

server:
  http:
    addr: localhost:8080
    tls:
      enable: true
      server_cert_path: ` + cert + `
      server_key_path: ` + key + `
  conns:
    read_timeout: 1s
    write_timeout: 1s
    idle_timeout: 1s

persistence:
  migrations_path: ./migrations
  postgres:
    host: localhost
    port: 5432
    sslmode: disable
    auth:
      user: postgres
      password: postgres
      db: notes
    conns:
      max_idles: 5
      max_opens: 10
      max_idle_time: 1m
      max_lifetime: 5m

cache:
  note_ttl: 5m
  redis:
    host: localhost
    port: 6379
`

	require.NoError(
		t,
		os.WriteFile(cfgPath, []byte(cfg), 0644),
	)

	result, err := New(cfgPath)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "notes", result.App.Name)
	assert.Equal(t, "1.0.0", result.App.Version)
	assert.Equal(t, "dev", result.App.Env)
}

func TestNew_InvalidConfig(t *testing.T) {
	dir := t.TempDir()

	cfgPath := filepath.Join(dir, "config.yaml")

	require.NoError(
		t,
		os.WriteFile(cfgPath, []byte("app: ["), 0644),
	)

	_, err := New(cfgPath)

	require.Error(t, err)
}
