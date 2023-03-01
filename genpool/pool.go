package main

import (
	"github.com/ergo-services/ergo/etf"
	"github.com/ergo-services/ergo/gen"
)

type MyPool struct {
	gen.Pool
}

func (p *MyPool) InitPool(process *gen.PoolProcess, args ...etf.Term) (gen.PoolOptions, error) {
	opts := gen.PoolOptions{
		Worker:     &MyPoolWorker{},
		NumWorkers: 5,
	}

	return opts, nil
}
