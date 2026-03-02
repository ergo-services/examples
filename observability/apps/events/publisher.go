package events

import (
	"fmt"
	"math/rand"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type publisher struct {
	act.Actor

	index int
	token gen.Ref
	name  gen.Atom
	seq   int
}

func factoryPublisher() gen.ProcessBehavior {
	return &publisher{}
}

func (p *publisher) Init(args ...any) error {
	p.index = args[0].(int)
	p.name = gen.Atom(fmt.Sprintf("evt_%d", p.index))

	var opts gen.EventOptions
	// 15-16: on_demand publishers
	if p.index >= 15 && p.index <= 16 {
		opts.Notify = true
	}

	token, err := p.RegisterEvent(p.name, opts)
	if err != nil {
		return err
	}
	p.token = token

	// schedule publishing based on category
	switch {
	case p.index <= 9:
		// active: publish every 1-5s
		p.schedulePublish()
	case p.index >= 12 && p.index <= 14:
		// no_subscribers: publish but nobody listens
		p.schedulePublish()
	}
	// 10-11: idle (no publish, no subscribers)
	// 15-16: on_demand (publish only when notified)
	// 17-19: no_publishing (never publish, will have subscribers)

	return nil
}

func (p *publisher) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case messagePublish:
		p.seq++
		p.SendEvent(p.name, p.token, MessageEventData{Seq: p.seq})
		p.Log().Debug("published %s seq %d", p.name, p.seq)
		p.schedulePublish()

	case gen.MessageEventStart:
		// on_demand: first subscriber appeared, start publishing
		if p.index >= 15 && p.index <= 16 {
			p.schedulePublish()
		}
	}
	return nil
}

func (p *publisher) schedulePublish() {
	delay := time.Duration(1+rand.Intn(5)) * time.Second
	p.SendAfter(p.PID(), messagePublish{}, delay)
}

func (p *publisher) Terminate(reason error) {}
