package node

import (
	"fmt"
	"nova/console"
	configenv "nova/core/env"
	"nova/log"
	"nova/service/account"
	"nova/service/gateway"
	"nova/service/resource"
	"nova/service/zone"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"ergo.services/ergo"
	"ergo.services/ergo/gen"
	"go.uber.org/zap"
)

var (
	configPath = filepath.Clean(configenv.DefaultConfigFile)

	startRequested bool
	stopRequested  bool
	checkRequested bool

	stateMu     sync.RWMutex
	runningNode gen.Node
)

func init() {
	console.RegisterBoolCommand("start", false, "-start", startNode)
	console.RegisterBoolCommand("stop", false, "-stop", stopNode)
	console.RegisterBoolCommand("check-config", false, "-check-config", checkConfig)
	console.RegisterStringCommand("config", configPath, "-config=./env.yaml", setConfigPath)
}

// Start parses all command-line options before initializing the logger and node.
// Command order therefore does not affect the final startup configuration.
func Start() error {
	if err := console.Run(os.Args); err != nil {
		return fmt.Errorf("parse startup options: %w", err)
	}

	if stopRequested {
		return Stop()
	}

	if !startRequested && !checkRequested {
		return nil
	}

	loader := configenv.Env{ConfigFile: configPath}
	if err := loader.Start(); err != nil {
		return fmt.Errorf("load service configuration: %w", err)
	}
	config, ok := configenv.Current()
	if !ok {
		return fmt.Errorf("load service configuration: no active configuration")
	}
	if checkRequested && !startRequested {
		return nil
	}

	fileConfig := loggerFileConfig(config.Logger)
	if err := log.Init(
		log.WithFileConfig(fileConfig),
		log.WithOption(zap.AddCallerSkip(1)),
	); err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}

	return startConfiguredNode(config, fileConfig)
}

func startConfiguredNode(config configenv.NodeConfig, fileConfig log.FileConfig) error {
	stateMu.Lock()
	defer stateMu.Unlock()

	if runningNode != nil && runningNode.IsAlive() {
		return fmt.Errorf("node %s is already running", runningNode.Name())
	}

	options := nodeOptions(config, fileConfig)
	nodeName := gen.Atom(config.NodeName)

	n, err := ergo.StartNode(nodeName, options)
	if err != nil {
		return fmt.Errorf("start node %s: %w", nodeName, err)
	}

	runningNode = n
	log.Info("node started success.")
	return nil
}

func Stop() error {
	stateMu.Lock()
	n := runningNode
	if n == nil || !n.IsAlive() {
		runningNode = nil
		stateMu.Unlock()
		return fmt.Errorf("node is not running")
	}
	runningNode = nil
	stateMu.Unlock()

	n.Stop()
	return nil
}

func IsRunning() bool {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return runningNode != nil && runningNode.IsAlive()
}

func Node() gen.Node {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return runningNode
}

func ConfigPath() string {
	return configPath
}

func startNode(value any) error {
	start, ok := value.(bool)
	if !ok {
		return fmt.Errorf("start must be a boolean")
	}
	if start && stopRequested {
		return fmt.Errorf("-start and -stop cannot be used together")
	}
	startRequested = start
	return nil
}

func stopNode(value any) error {
	stop, ok := value.(bool)
	if !ok {
		return fmt.Errorf("stop must be a boolean")
	}
	if stop && startRequested {
		return fmt.Errorf("-start and -stop cannot be used together")
	}
	stopRequested = stop
	return nil
}

func checkConfig(value any) error {
	check, ok := value.(bool)
	if !ok {
		return fmt.Errorf("check-config must be a boolean")
	}
	checkRequested = check
	return nil
}

func ergoLogLevel(level string) gen.LogLevel {
	switch level {
	case "debug":
		return gen.LogLevelDebug
	case "warn":
		return gen.LogLevelWarning
	case "error":
		return gen.LogLevelError
	case "panic", "fatal":
		return gen.LogLevelPanic
	default:
		return gen.LogLevelInfo
	}
}

func setConfigPath(value any) error {
	path, ok := value.(string)
	if !ok {
		return fmt.Errorf("config path must be a string")
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("config path cannot be empty")
	}
	configPath = filepath.Clean(path)
	return nil
}

func loggerFileConfig(config configenv.LoggerConfig) log.FileConfig {
	logPath, logFileName := filepath.Split(filepath.Clean(config.FileName))
	if logPath == "" {
		logPath = "."
	}
	return log.FileConfig{
		LogLevel:      strings.ToLower(config.Level),
		LogPath:       filepath.Clean(logPath),
		LogFileName:   logFileName,
		LogMaxSize:    config.MaxSize,
		LogMaxBackups: config.MaxBackups,
		LogMaxAge:     config.MaxAge,
		Compress:      config.Compress,
	}
}

func nodeOptions(config configenv.NodeConfig, fileConfig log.FileConfig) gen.NodeOptions {
	options := gen.NodeOptions{
		Network: gen.NetworkOptions{
			Mode: gen.NetworkModeEnabled, Cookie: config.Cookie,
			MaxMessageSize: 1 << 20,
			Flags: func() gen.NetworkFlags {
				flags := gen.DefaultNetworkFlags
				flags.EnableImportantDelivery = true
				return flags
			}(),
			Acceptors: []gen.AcceptorOptions{
				{
					Host:      config.Host,
					Port:      uint16(config.Port),
					PortRange: 1,
				},
			},
		},
		Log: gen.LogOptions{
			Level: ergoLogLevel(fileConfig.LogLevel),
			DefaultLogger: gen.DefaultLoggerOptions{
				Disable: true,
			},
			Loggers: []gen.Logger{
				{
					Name:   "zap-logger",
					Logger: log.CreateLogger(),
				},
			},
		},
	}
	switch config.NodeType {
	case "zc":
		resources := resource.NewApplication(config, resource.Options{MySQL: true, Redis: true})
		options.Applications = []gen.ApplicationBehavior{resources, account.NewApplication(config, resources)}
	case "gate":
		options.Applications = []gen.ApplicationBehavior{gateway.NewApplication(config)}
	case "zone":
		resources := resource.NewApplication(config, resource.Options{MySQL: true})
		options.Applications = []gen.ApplicationBehavior{resources, zone.NewApplication(config, resources)}
	}
	return options
}
