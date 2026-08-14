package main

import (
	"fmt"
	"nova/log"
	"nova/node"
	"os"
)

func main() {
	if err := node.Start(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !node.IsRunning() {
		return
	}
	defer func() { _ = log.Sync() }()

	node.Node().Wait()

}
