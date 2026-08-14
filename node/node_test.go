package node

import (
	configenv "nova/core/env"
	"path/filepath"
	"testing"

	"ergo.services/ergo/gen"
)

func TestLoggerFileConfig(t *testing.T) {
	got := loggerFileConfig(configenv.LoggerConfig{
		Level: "debug", FileName: "./syslog/console.log",
		MaxSize: 2, MaxBackups: 10, MaxAge: 7, Compress: true,
	})
	if got.LogPath != "syslog" || got.LogFileName != "console.log" {
		t.Fatalf("unexpected log target: path=%q file=%q", got.LogPath, got.LogFileName)
	}
	if got.LogLevel != "debug" || got.LogMaxSize != 2 || !got.Compress {
		t.Fatalf("unexpected log configuration: %+v", got)
	}
}

func TestNodeOptionsComeFromServiceConfig(t *testing.T) {
	config := configenv.NodeConfig{
		Host: "127.0.0.1", Port: 50012, Cookie: "cookie-from-yaml",
	}
	fileConfig := loggerFileConfig(configenv.LoggerConfig{
		Level: "warn", FileName: "nova.log", MaxSize: 1,
	})
	options := nodeOptions(config, fileConfig)

	if options.Network.Cookie != "cookie-from-yaml" {
		t.Fatalf("cookie = %q", options.Network.Cookie)
	}
	if len(options.Network.Acceptors) != 1 {
		t.Fatalf("acceptors = %d", len(options.Network.Acceptors))
	}
	acceptor := options.Network.Acceptors[0]
	if acceptor.Host != "127.0.0.1" || acceptor.Port != 50012 || acceptor.PortRange != 1 {
		t.Fatalf("unexpected acceptor: %+v", acceptor)
	}
	if options.Log.Level != gen.LogLevelWarning {
		t.Fatalf("log level = %v", options.Log.Level)
	}
}

func TestAccountApplicationOnlyLoadsOnZCNode(t *testing.T) {
	config := configenv.NodeConfig{Host: "127.0.0.1", Port: 50012, Cookie: "cookie", NodeType: "zc"}
	options := nodeOptions(config, loggerFileConfig(configenv.LoggerConfig{Level: "info", FileName: "nova.log", MaxSize: 1}))
	if len(options.Applications) != 2 {
		t.Fatalf("zc applications = %d, want 2", len(options.Applications))
	}

	config.NodeType = "zone"
	options = nodeOptions(config, loggerFileConfig(configenv.LoggerConfig{Level: "info", FileName: "nova.log", MaxSize: 1}))
	if len(options.Applications) != 2 {
		t.Fatalf("zone applications = %d, want 2", len(options.Applications))
	}

	config.NodeType = "gate"
	options = nodeOptions(config, loggerFileConfig(configenv.LoggerConfig{Level: "info", FileName: "nova.log", MaxSize: 1}))
	if len(options.Applications) != 1 {
		t.Fatalf("gate applications = %d, want 1", len(options.Applications))
	}
}

func TestSetConfigPath(t *testing.T) {
	previous := configPath
	t.Cleanup(func() { configPath = previous })

	if err := setConfigPath("/etc/nova/env.yaml"); err != nil {
		t.Fatalf("setConfigPath() returned error: %v", err)
	}
	if configPath != filepath.Clean("/etc/nova/env.yaml") {
		t.Fatalf("configPath = %q", configPath)
	}
}

func TestCheckConfigSetter(t *testing.T) {
	previous := checkRequested
	t.Cleanup(func() { checkRequested = previous })

	if err := checkConfig(true); err != nil {
		t.Fatalf("checkConfig() returned error: %v", err)
	}
	if !checkRequested {
		t.Fatal("checkRequested = false")
	}
}
