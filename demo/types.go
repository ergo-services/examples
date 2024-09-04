package demo

import (
	"ergo.services/ergo/gen"
	"ergo.services/ergo/net/edf"
)

type MyMsg1 struct {
	// Add your fields
}
type MyMsg2 struct {
	// Add your fields
}

func init() {
	types := []any{
		MyMsg1{},
		MyMsg2{},
	}

	for _, t := range types {
		err := edf.RegisterTypeOf(t)
		if err == nil || err == gen.ErrTaken {
			continue
		}
		panic(err)
	}
}
