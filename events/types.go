package main

import (
	"ergo.services/ergo/net/edf"
)

type MyPubMessage struct {
	MyString string
}

func init() {
	// register network messages
	if err := edf.RegisterTypeOf(MyPubMessage{}); err != nil {
		panic(err)
	}
}
