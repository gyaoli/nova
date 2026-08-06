package node

import (
	"fmt"
	"nova/console"
	"nova/core/log"
	"os"

	"go.uber.org/zap"
	"ergo.services/ergo"
)

var fileConfig = log.FileConfig{}

func init() {
	console.RegisterStringCommand("name", "gameserver", "-name=gameserver", setNodeName)
	console.RegisterBoolCommand("start", false, "-start=true", startNode)
	console.RegisterBoolCommand("stop", false, "-stop=true", stopNode)
	console.RegisterStringCommand("config", "", "-config=./config", setConfigPath)
	console.RegisterStringCommand("logLevel", "debug", "-logLevel=debug", setLogLevel)
	console.RegisterStringCommand("logPath", "./", "-logPath=./logDir", setLogPath)

}

func Start() {
	err := console.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	_ = log.Init(log.WithFileConfig(fileConfig), log.WithConsole(), log.WithOption(zap.AddCallerSkip(1)))

}

func setNodeName(args interface{}) error {

	return nil
}

func startNode(node interface{}) error {
	return nil
}

func stopNode(args interface{}) error {
	return nil
}

func setLogLevel(level interface{}) error {
	fileConfig.LogLevel = level.(string)
	return nil
}

func setLogPath(path interface{}) error {
	fmt.Println(path)
	fileConfig.LogPath = path.(string)
	return nil
}

func setConfigPath(args interface{}) error {
	return nil
}
