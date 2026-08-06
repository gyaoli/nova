package main

import (
	"fmt"
	"nova/core/actor"
	"nova/logo"
	"nova/node"

	console "github.com/asynkron/goconsole"
)

type sayHello struct {
	actor.ActorBase
}

func (s sayHello) OnStart(ctx actor.Context) {

}

func (s sayHello) OnhandleSend(ctx actor.Context, massage interface{}) {
	fmt.Println(massage)
}

func (s sayHello) OnhandleCall(ctx actor.Context, message interface{}) (reply interface{}) {
	return nil

}

type sayBye struct {
	actor.ActorBase
}

func (s sayBye) OnhandleSend(ctx actor.Context, massage interface{}) {
	switch msg := massage.(type) {
	case string:
		fmt.Println(msg)
		ctx.Send(ctx.Sender(), "yaa,i am bye")
	}
}

func (s sayBye) OnhandleCall(ctx actor.Context, message interface{}) (reply interface{}) {
	return nil

}

func main() {
	logo.Print()
	node.Start()

	hello, err := actor.Spawn(&sayHello{})
	if err != nil {
		return
	}

	bye, err := actor.Spawn(&sayBye{})
	if err != nil {
		return
	}

	_ = hello.Send(bye.Pid(), "hi! i am hello")

	_, _ = console.ReadLine()

}
