package main

import (
	"fmt"

	"github.com/ergo-services/ergo/etf"
	"github.com/ergo-services/ergo/gen"
)

type MyPoolWorker struct {
	gen.PoolWorker
}

func (mpw *MyPoolWorker) InitPoolWorker(process *gen.PoolWorkerProcess, args ...etf.Term) error {
	fmt.Println("   started pool worker: ", process.Self())
	return nil
}

func (mpw *MyPoolWorker) HandleWorkerCall(process *gen.PoolWorkerProcess, message etf.Term) etf.Term {
	fmt.Printf("[%s] received Call request: %v\n", process.Self(), message)
	return "pong"
}

func (mpw *MyPoolWorker) HandleWorkerCast(process *gen.PoolWorkerProcess, message etf.Term) {
	fmt.Printf("[%s] received Cast message: %v\n", process.Self(), message)
}

func (mpw *MyPoolWorker) HandleWorkerInfo(process *gen.PoolWorkerProcess, message etf.Term) {
	fmt.Printf("[%s] received Info message: %v\n", process.Self(), message)
}
