package main

import (
	"flag"
	"fmt"

	"github.com/ergo-services/ergo"
	"github.com/ergo-services/ergo/gen"
	"github.com/ergo-services/ergo/node"
)

const (
	poolProcessName = "mypool"
)

func main() {
	flag.Parse()

	fmt.Printf("Starting node: pool@localhost...")
	node1, err := ergo.StartNode("pool@localhost", "cookies", node.Options{})
	if err != nil {
		panic(err)
	}
	fmt.Println("OK")

	fmt.Println("Starting mypool process ...")
	pool := &MyPool{}
	if _, err := node1.Spawn(poolProcessName, gen.ProcessOptions{}, pool); err != nil {
		panic(err)
	}
	fmt.Println("OK")

	fmt.Printf("Starting myping process ...")
	ping := &MyPing{}
	if _, err := node1.Spawn("myping", gen.ProcessOptions{}, ping); err != nil {
		panic(err)
	}
	fmt.Println("OK")

	node1.Wait()
}
