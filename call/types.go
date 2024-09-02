package main

import (
	"ergo.services/ergo/net/edf"
)

type MyRequest struct {
	MyBool   bool
	MyString string
}

func init() {
	// register network messages
	if err := edf.RegisterTypeOf(MyRequest{}); err != nil {
		panic(err)
	}
}
