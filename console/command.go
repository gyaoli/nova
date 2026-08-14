package console

import (
	"flag"
	"fmt"
)

type valueType int8
type CommandFunc func(args interface{}) error

var CommandList []*Command

const (
	CommandValBool   valueType = 1
	CommandValInt    valueType = 2
	CommandValString valueType = 3
)

type Command struct {
	name    string
	valType valueType
	usage   string
	boolVal bool
	intVal  int
	strVal  string
	fn      CommandFunc
}

func RegisterBoolCommand(name string, defaultValue bool, usage string, fn CommandFunc) {
	c := &Command{name: name, valType: CommandValBool, usage: usage, fn: fn}
	flag.BoolVar(&c.boolVal, name, defaultValue, usage)
	CommandList = append(CommandList, c)
}

func RegisterIntCommand(name string, defaultValue int, usage string, fn CommandFunc) {
	c := &Command{name: name, valType: CommandValInt, usage: usage, fn: fn}
	flag.IntVar(&c.intVal, name, defaultValue, usage)
	CommandList = append(CommandList, c)
}

func RegisterStringCommand(name string, defaultValue string, usage string, fn CommandFunc) {
	c := &Command{name: name, valType: CommandValString, usage: usage, fn: fn}
	flag.StringVar(&c.strVal, name, defaultValue, usage)
	CommandList = append(CommandList, c)
}

func (c *Command) execute() error {
	switch c.valType {
	case CommandValBool:
		return c.fn(c.boolVal)
	case CommandValInt:
		return c.fn(c.intVal)
	case CommandValString:
		return c.fn(c.strVal)
	default:
		return fmt.Errorf("unknow command type")
	}
}

func Run(args []string) error {
	flag.Parse()
	for _, cmd := range CommandList {
		err := cmd.execute()
		if err != nil {
			return err
		}
	}

	return nil
}
