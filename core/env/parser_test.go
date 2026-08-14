package env

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadCurrentEnvYAML(t *testing.T) {
	config, err := Load(filepath.Join("..", "..", "env.yaml"))
	if err != nil {
		t.Fatalf("Load(env.yaml) returned error: %v", err)
	}

	if config.Version != "1.0" || config.GameCode != "nova" {
		t.Fatalf("unexpected identity fields: %+v", config)
	}
	if config.NetType != "tcp" || config.NodeName != "nova1@localhost" || config.Cookie != "123abc" {
		t.Fatalf("unexpected node fields: %+v", config)
	}
	if config.Logger.FileName != "./syslog/console.log" || config.Mysql.Port != 3306 {
		t.Fatalf("unexpected nested fields: %+v", config)
	}
	if config.Redis.Address != "127.0.0.1:6379" || config.Account.TokenTTL != 2*time.Hour {
		t.Fatalf("unexpected account dependencies: %+v", config)
	}
}

func TestLoadValidatesZCDependencies(t *testing.T) {
	content := strings.Replace(validConfigYAML, "node_type: zone", "node_type: zc", 1)
	path := writeConfig(t, content)
	if _, err := Load(path); err != nil {
		t.Fatalf("Load(zc config) returned error: %v", err)
	}

	content = strings.Replace(content, "address: 127.0.0.1:6379", "address: ''", 1)
	path = writeConfig(t, content)
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "redis.address") {
		t.Fatalf("Load(zc without redis) error=%v", err)
	}
}

func TestLoadGateDoesNotRequireStorageResources(t *testing.T) {
	content := strings.Replace(validConfigYAML, "node_type: zone", "node_type: gate", 1)
	content = strings.Replace(content, "web_port: 50013", "web_port: 0", 1)
	content = strings.Replace(content, "host: 127.0.0.1\n  port: 3306\n  user: root", "host: ''\n  port: 0\n  user: ''", 1)
	content = strings.Replace(content, "db: nova_dev_1", "db: ''", 1)
	content = strings.Replace(content, "address: 127.0.0.1:6379", "address: ''", 1)
	path := writeConfig(t, content)
	if _, err := Load(path); err != nil {
		t.Fatalf("Load(gate without storage resources) returned error: %v", err)
	}
}

func TestLoadRejectsUnknownField(t *testing.T) {
	content := validConfigYAML + "unknown_field: true\n"
	path := writeConfig(t, content)
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "unknown_field") {
		t.Fatalf("Load() error = %v, want unknown field error", err)
	}
}

func TestLoadRejectsInvalidConfig(t *testing.T) {
	path := writeConfig(t, strings.Replace(validConfigYAML, "port: 50012", "port: 0", 1))
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "port must be") {
		t.Fatalf("Load() error = %v, want invalid port error", err)
	}
}

func TestEnvStartPublishesSnapshot(t *testing.T) {
	path := writeConfig(t, validConfigYAML)
	e := Env{ConfigFile: path}
	if err := e.Start(); err != nil {
		t.Fatalf("Start() returned error: %v", err)
	}
	config, ok := Current()
	if !ok || config.NodeID != 1 {
		t.Fatalf("Current() = (%+v, %v)", config, ok)
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "env.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

const validConfigYAML = `version: "1.0"
game_code: nova
local_area: cn
lang: zh_CN
host: 127.0.0.1
net_type: tcp
port: 50012
web_port: 50013
platform: dev
node_id: 1
node_name: nova1@localhost
cookie: test-cookie
node_type: zone
logger:
  level: debug
  filename: ./syslog/console.log
  max_size: 2
  max_backups: 10
  max_age: 10
  compress: false
mysql:
  host: 127.0.0.1
  port: 3306
  user: root
  password: test-password
  db: nova_dev_1
  max_open_conns: 32
  max_idle_conns: 16
  conn_max_lifetime: 30m
  conn_max_idle_time: 5m
redis:
  address: 127.0.0.1:6379
  username: ""
  password: ""
  db: 0
  pool_size: 32
  dial_timeout: 2s
  read_timeout: 1s
  write_timeout: 1s
account:
  token_ttl: 2h
  http_max_in_flight: 128
gateway:
  listen_host: 127.0.0.1
  listen_port: 50014
  auth_url: http://127.0.0.1:50013/account/verify
  zone_node: zone1@localhost
  zone_host: 127.0.0.1
  zone_port: 50022
zone:
  id: 1
`
