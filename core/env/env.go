package env

import (
	"nova/core/module"
)

type Env struct {
	module.IModule
}

type NodeConfig struct {
	Version   float32 `mapstructure:"version" json:"version"`
	GameCode  string  `mapstructure:"game_code" json:"game_code"`
	LocalArea string  `mapstructure:"local_area" json:"local_area"`
	Lang      string  `mapstructure:"lang" json:"lang"`
	Host      string  `mapstructure:"host" json:"host"`
	Port      int     `mapstructure:"port" json:"port"`
	WebPort   int     `mapstructure:"web_port" json:"web_port"`
	NodeId    int     `mapstructure:"node_id" json:"node_id"`
	NodeType  string  `mapstructure:"node_type" json:"node_type"`
	NodeName  string  `mapstructure:"node_name" json:"node_name"`
	Platform  string  `mapstructure:"platform" json:"platform"`

	Mysql  MysqlConfig  `mapstructure:"mysql" json:"mysql"`
	Logger LoggerConfig `mapstructure:"logger" json:"logger"`
}

type MysqlConfig struct {
	User     string `mapstructure:"user" json:"user"`
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	DB       string `mapstructure:"db" json:"db"`
	PassWord string `mapstructure:"password" json:"password"`
}

type LoggerConfig struct {
	Level      string `mapstructure:"level" json:"level"`
	FileName   string `mapstructure:"filename" json:"filename"`
	MaxSize    int    `mapstructure:"max_size" json:"max_size"`
	MaxBackups int    `mapstructure:"max_backups" json:"max_backups"`
	MaxAge     int    `mapstructure:"max_age" json:"max_age"`
	Compress   bool   `mapstructure:"compress" json:"compress"`
}



func (env *Env) Init() error {
	return nil
}
func (env *Env) Start() error{
	parser()
	return nil
}

func (env *Env) OnAfterStart() {

}

func (env *Env) OnBeforeStop() {
}

func (env *Env) OnStop() {
}
