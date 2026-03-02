package lifecycle

import (
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type zombieMaker struct {
	act.Actor
}

type messageBlock struct{}
type messageSpawnZombie struct{}
type messageKillZombie struct {
	pid gen.PID
}

func factoryZombieMaker() gen.ProcessBehavior {
	return &zombieMaker{}
}

func (z *zombieMaker) Init(args ...any) error {
	z.Log().Info("zombie maker started on %s", z.Node().Name())
	z.SendAfter(z.PID(), messageSpawnZombie{}, 5*time.Second)
	return nil
}

func (z *zombieMaker) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case messageSpawnZombie:
		pid, err := z.Spawn(factoryZombieChild, gen.ProcessOptions{})
		if err != nil {
			z.Log().Warning("zombie maker: spawn failed: %s", err)
			return nil
		}
		z.Log().Info("zombie maker: spawned zombie child %s, sending block message", pid)
		z.Send(pid, messageBlock{})
		// give it a moment to enter the blocking select{}, then kill
		z.SendAfter(z.PID(), messageKillZombie{pid: pid}, 2*time.Second)

	case messageKillZombie:
		m := message.(messageKillZombie)
		z.Log().Error("zombie maker: child %s is not responding", m.pid)
		z.Log().Warning("zombie maker: killing blocked child %s", m.pid)
		z.Node().Kill(m.pid)
	}
	return nil
}

func (z *zombieMaker) Terminate(reason error) {}

// zombieChild blocks forever in HandleMessage, becoming a zombie when killed
type zombieChild struct {
	act.Actor
}

func factoryZombieChild() gen.ProcessBehavior {
	return &zombieChild{}
}

func (c *zombieChild) Init(args ...any) error {
	return nil
}

func (c *zombieChild) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case messageBlock:
		c.processPayloadDecompression()
	}
	return nil
}

//go:noinline
func (c *zombieChild) processPayloadDecompression() {
	select {}
}

func (c *zombieChild) Terminate(reason error) {}
