package lifecycle

import (
	"fmt"
	"math/rand"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type child struct {
	act.Actor
}

func factoryChild() gen.ProcessBehavior {
	return &child{}
}

func (c *child) Init(args ...any) error {
	// random init delay 0..1s
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

	ttl := time.Duration(1+rand.Intn(30)) * time.Second // 1..30s
	c.SendAfter(c.PID(), messageDie{}, ttl)
	return nil
}

func (c *child) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case messageDie:
		c.Log().Warning("lifecycle child %s: terminating", c.PID())
		return fmt.Errorf("lifecycle child: random termination")
	}
	return nil
}

func (c *child) Terminate(reason error) {
	c.Log().Error("lifecycle child %s terminated: %s", c.PID(), reason)
}
