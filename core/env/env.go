package env

import (
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"time"
)

const DefaultConfigFile = "./env.yaml"

var current atomic.Pointer[NodeConfig]

// Env loads the service configuration during application startup.
type Env struct {
	ConfigFile string
}

// NodeConfig mirrors env.yaml. A loaded configuration is immutable; callers
// receive a value snapshot through Current rather than a shared mutable pointer.
type NodeConfig struct {
	Version   string        `mapstructure:"version" json:"version" yaml:"version"`
	GameCode  string        `mapstructure:"game_code" json:"game_code" yaml:"game_code"`
	LocalArea string        `mapstructure:"local_area" json:"local_area" yaml:"local_area"`
	Lang      string        `mapstructure:"lang" json:"lang" yaml:"lang"`
	Host      string        `mapstructure:"host" json:"host" yaml:"host"`
	NetType   string        `mapstructure:"net_type" json:"net_type" yaml:"net_type"`
	Port      int           `mapstructure:"port" json:"port" yaml:"port"`
	WebPort   int           `mapstructure:"web_port" json:"web_port" yaml:"web_port"`
	Platform  string        `mapstructure:"platform" json:"platform" yaml:"platform"`
	NodeID    int           `mapstructure:"node_id" json:"node_id" yaml:"node_id"`
	NodeName  string        `mapstructure:"node_name" json:"node_name" yaml:"node_name"`
	Cookie    string        `mapstructure:"cookie" json:"cookie" yaml:"cookie"`
	NodeType  string        `mapstructure:"node_type" json:"node_type" yaml:"node_type"`
	Mysql     MysqlConfig   `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Redis     RedisConfig   `mapstructure:"redis" json:"redis" yaml:"redis"`
	Account   AccountConfig `mapstructure:"account" json:"account" yaml:"account"`
	Gateway   GatewayConfig `mapstructure:"gateway" json:"gateway" yaml:"gateway"`
	Zone      ZoneConfig    `mapstructure:"zone" json:"zone" yaml:"zone"`
	Logger    LoggerConfig  `mapstructure:"logger" json:"logger" yaml:"logger"`
}

type GatewayConfig struct {
	ListenHost string `mapstructure:"listen_host" json:"listen_host" yaml:"listen_host"`
	ListenPort int    `mapstructure:"listen_port" json:"listen_port" yaml:"listen_port"`
	AuthURL    string `mapstructure:"auth_url" json:"auth_url" yaml:"auth_url"`
	ZoneNode   string `mapstructure:"zone_node" json:"zone_node" yaml:"zone_node"`
	ZoneHost   string `mapstructure:"zone_host" json:"zone_host" yaml:"zone_host"`
	ZonePort   int    `mapstructure:"zone_port" json:"zone_port" yaml:"zone_port"`
}

type ZoneConfig struct {
	ID int `mapstructure:"id" json:"id" yaml:"id"`
}

type MysqlConfig struct {
	User            string        `mapstructure:"user" json:"user" yaml:"user"`
	Host            string        `mapstructure:"host" json:"host" yaml:"host"`
	Port            int           `mapstructure:"port" json:"port" yaml:"port"`
	DB              string        `mapstructure:"db" json:"db" yaml:"db"`
	Password        string        `mapstructure:"password" json:"password" yaml:"password"`
	MaxOpenConns    int           `mapstructure:"max_open_conns" json:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" json:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time" json:"conn_max_idle_time" yaml:"conn_max_idle_time"`
}

type RedisConfig struct {
	Address      string        `mapstructure:"address" json:"address" yaml:"address"`
	Username     string        `mapstructure:"username" json:"username" yaml:"username"`
	Password     string        `mapstructure:"password" json:"password" yaml:"password"`
	DB           int           `mapstructure:"db" json:"db" yaml:"db"`
	PoolSize     int           `mapstructure:"pool_size" json:"pool_size" yaml:"pool_size"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout" json:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" json:"write_timeout" yaml:"write_timeout"`
}

type AccountConfig struct {
	TokenTTL        time.Duration `mapstructure:"token_ttl" json:"token_ttl" yaml:"token_ttl"`
	HTTPMaxInFlight int           `mapstructure:"http_max_in_flight" json:"http_max_in_flight" yaml:"http_max_in_flight"`
}

type LoggerConfig struct {
	Level      string `mapstructure:"level" json:"level" yaml:"level"`
	FileName   string `mapstructure:"filename" json:"filename" yaml:"filename"`
	MaxSize    int    `mapstructure:"max_size" json:"max_size" yaml:"max_size"`
	MaxBackups int    `mapstructure:"max_backups" json:"max_backups" yaml:"max_backups"`
	MaxAge     int    `mapstructure:"max_age" json:"max_age" yaml:"max_age"`
	Compress   bool   `mapstructure:"compress" json:"compress" yaml:"compress"`
}

func (e *Env) Init() error { return nil }

func (e *Env) Start() error {
	path := e.ConfigFile
	if strings.TrimSpace(path) == "" {
		path = DefaultConfigFile
	}

	config, err := Load(path)
	if err != nil {
		return err
	}
	publish(config)
	return nil
}

// Current returns a copy of the active configuration.
func Current() (NodeConfig, bool) {
	config := current.Load()
	if config == nil {
		return NodeConfig{}, false
	}
	return *config, true
}

func publish(config NodeConfig) {
	snapshot := config
	current.Store(&snapshot)
}

func (config NodeConfig) validate() error {
	required := map[string]string{
		"version": config.Version, "game_code": config.GameCode,
		"local_area": config.LocalArea, "lang": config.Lang,
		"host": config.Host, "net_type": config.NetType,
		"platform": config.Platform, "node_name": config.NodeName,
		"cookie": config.Cookie, "node_type": config.NodeType,
		"logger.level": config.Logger.Level, "logger.filename": config.Logger.FileName,
	}
	for field, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s cannot be empty", field)
		}
	}

	if net.ParseIP(config.Host) == nil && strings.ContainsAny(config.Host, " /\\") {
		return fmt.Errorf("host %q is invalid", config.Host)
	}
	if config.NetType != "tcp" && config.NetType != "ws" {
		return fmt.Errorf("net_type %q is invalid: expected tcp or ws", config.NetType)
	}
	if config.NodeType != "zc" && config.NodeType != "gate" && config.NodeType != "zone" {
		return fmt.Errorf("node_type %q is invalid: expected zc, gate, or zone", config.NodeType)
	}
	if config.NodeType == "zc" {
		if err := config.validateMySQL(); err != nil {
			return err
		}
		if config.WebPort < 1 || config.WebPort > 65535 {
			return fmt.Errorf("web_port must be between 1 and 65535")
		}
		if config.Port == config.WebPort {
			return fmt.Errorf("port and web_port cannot be the same")
		}
		if strings.TrimSpace(config.Redis.Address) == "" {
			return fmt.Errorf("redis.address cannot be empty for zc node")
		}
		if config.Redis.DB < 0 {
			return fmt.Errorf("redis.db cannot be negative")
		}
		if config.Account.TokenTTL <= 0 {
			return fmt.Errorf("account.token_ttl must be greater than zero")
		}
		if config.Account.HTTPMaxInFlight <= 0 {
			return fmt.Errorf("account.http_max_in_flight must be greater than zero")
		}
	}
	if config.NodeType == "gate" {
		if config.NetType != "tcp" {
			return fmt.Errorf("net_type %q is not implemented for gate node: expected tcp", config.NetType)
		}
		if strings.TrimSpace(config.Gateway.ListenHost) == "" || strings.TrimSpace(config.Gateway.AuthURL) == "" || strings.TrimSpace(config.Gateway.ZoneNode) == "" || strings.TrimSpace(config.Gateway.ZoneHost) == "" {
			return fmt.Errorf("gateway listen_host, auth_url, zone_node, and zone_host cannot be empty for gate node")
		}
		for field, port := range map[string]int{"gateway.listen_port": config.Gateway.ListenPort, "gateway.zone_port": config.Gateway.ZonePort} {
			if port < 1 || port > 65535 {
				return fmt.Errorf("%s must be between 1 and 65535", field)
			}
		}
		if config.Gateway.ListenPort == config.Port {
			return fmt.Errorf("gateway.listen_port and port cannot be the same")
		}
	}
	if config.NodeType == "zone" {
		if err := config.validateMySQL(); err != nil {
			return err
		}
		if config.Zone.ID <= 0 {
			return fmt.Errorf("zone.id must be greater than zero for zone node")
		}
	}
	if config.NodeID <= 0 {
		return fmt.Errorf("node_id must be greater than zero")
	}
	if parts := strings.Split(config.NodeName, "@"); len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("node_name %q is invalid: expected name@host", config.NodeName)
	}
	for field, port := range map[string]int{"port": config.Port} {
		if port < 1 || port > 65535 {
			return fmt.Errorf("%s must be between 1 and 65535", field)
		}
	}
	switch strings.ToLower(config.Logger.Level) {
	case "debug", "info", "warn", "error", "panic", "fatal":
	default:
		return fmt.Errorf("logger.level %q is invalid", config.Logger.Level)
	}
	if config.Logger.MaxSize <= 0 || config.Logger.MaxBackups < 0 || config.Logger.MaxAge < 0 {
		return fmt.Errorf("logger rotation values are invalid")
	}
	return nil
}

func (config NodeConfig) validateMySQL() error {
	for field, value := range map[string]string{
		"mysql.host": config.Mysql.Host,
		"mysql.user": config.Mysql.User,
		"mysql.db":   config.Mysql.DB,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s cannot be empty", field)
		}
	}
	if config.Mysql.Port < 1 || config.Mysql.Port > 65535 {
		return fmt.Errorf("mysql.port must be between 1 and 65535")
	}
	return nil
}
