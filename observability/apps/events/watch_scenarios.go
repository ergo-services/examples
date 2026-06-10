package events

import (
	"fmt"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

// Watch-window demo events. Open the observer "event" window for each to see a
// different watch banner. Payloads are realistic structs (slices/maps) to exercise
// the detail pretty-printer; the inspector stringifies them locally, so no EDF
// registration is needed.
//
//   temperature_sensor  Notify=true, samples only while subscribed → "idle_gated" (passive)
//                       + "Watch anyway" (forced)
//   chat_room           gated, with a member that joins/leaves every ~12s →
//                       flips "other_subscribers" ↔ "idle_gated" live
//   orders_created      Notify=true, fires on every order regardless → "publishes_regardless"
//   service_heartbeat   Notify=false, ticks always → "notify_off"
//   audit_trail         Notify=false, ~200 msg/s → lower the window limit to see "suppressed"
//   deploy_status       Notify=false, registered per deploy then unregistered → "closed"

const (
	evTemperature gen.Atom = "temperature_sensor"
	evChat        gen.Atom = "chat_room"
	evOrders      gen.Atom = "orders_created"
	evHeartbeat   gen.Atom = "service_heartbeat"
	evAudit       gen.Atom = "audit_trail"
	evDeploy      gen.Atom = "deploy_status"
)

type TemperatureReading struct {
	Sensor  string
	Celsius float64
	Tags    []string
	Meta    map[string]int
}

type OrderLine struct {
	SKU   string
	Qty   int
	Price float64
}

type OrderCreated struct {
	ID       int
	Customer string
	Items    []OrderLine
	Total    float64
}

type Heartbeat struct {
	Service   string
	Seq       int
	UptimeSec int
}

type ChatMessage struct {
	Room string
	From string
	Text string
}

type AuditEntry struct {
	Seq    int
	Action string
	Actor  string
}

type DeployStatus struct {
	Release  string
	Phase    string
	Progress int
}

type watchTick struct{ id uint64 }
type msgRegister struct{}

// register the watch demo events after the load set so they sort to the top (newest first)
const watchRegisterDelay = 1 * time.Second

// samplePayload builds the realistic payload for a given demo event.
func samplePayload(name gen.Atom, seq int) any {
	switch name {
	case evChat:
		return ChatMessage{Room: "room42", From: fmt.Sprintf("user%d", seq%5), Text: fmt.Sprintf("message #%d", seq)}
	case evOrders:
		return OrderCreated{
			ID:       1000 + seq,
			Customer: fmt.Sprintf("acme-%d", seq%7),
			Items:    []OrderLine{{SKU: "SKU-A", Qty: seq%3 + 1, Price: 9.99}, {SKU: "SKU-B", Qty: 1, Price: 19.50}},
			Total:    39.49,
		}
	case evHeartbeat:
		return Heartbeat{Service: "api-gateway", Seq: seq, UptimeSec: seq}
	default: // temperature_sensor
		return TemperatureReading{Sensor: "boiler-1", Celsius: 55 + float64(seq%40), Tags: []string{"hvac", "zone-a"}, Meta: map[string]int{"reading": seq}}
	}
}

// ── gated producer: publishes only between MessageEventStart and MessageEventStop ──

func factoryTemperatureSensor() gen.ProcessBehavior { return &gatedProducer{name: evTemperature} }
func factoryChatRoom() gen.ProcessBehavior        { return &gatedProducer{name: evChat} }

type gatedProducer struct {
	act.Actor
	name   gen.Atom
	token  gen.Ref
	active bool
	loopID uint64
	seq    int
}

func (p *gatedProducer) Init(args ...any) error {
	p.SendAfter(p.PID(), msgRegister{}, watchRegisterDelay)
	return nil
}

func (p *gatedProducer) HandleMessage(from gen.PID, message any) error {
	switch m := message.(type) {
	case msgRegister:
		token, err := p.RegisterEvent(p.name, gen.EventOptions{Notify: true, Buffer: 5})
		if err != nil {
			return err
		}
		p.token = token
	case gen.MessageEventStart:
		p.active = true
		p.loopID++
		p.Send(p.PID(), watchTick{p.loopID})
	case gen.MessageEventStop:
		p.active = false
	case watchTick:
		if m.id != p.loopID || p.active == false {
			break
		}
		p.seq++
		p.SendEvent(p.name, p.token, samplePayload(p.name, p.seq))
		p.SendAfter(p.PID(), watchTick{p.loopID}, time.Second)
	}
	return nil
}

func (p *gatedProducer) Terminate(reason error) { p.UnregisterEvent(p.name) }

// ── timer producer: publishes on an interval regardless of subscribers ──

func factoryOrdersCreated() gen.ProcessBehavior {
	return &timerProducer{name: evOrders, notify: true, interval: time.Second}
}
func factoryServiceHeartbeat() gen.ProcessBehavior {
	return &timerProducer{name: evHeartbeat, notify: false, interval: time.Second}
}

type timerProducer struct {
	act.Actor
	name     gen.Atom
	notify   bool
	interval time.Duration
	token    gen.Ref
	seq      int
}

func (p *timerProducer) Init(args ...any) error {
	p.SendAfter(p.PID(), msgRegister{}, watchRegisterDelay)
	return nil
}

func (p *timerProducer) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case msgRegister:
		token, err := p.RegisterEvent(p.name, gen.EventOptions{Notify: p.notify, Buffer: 5})
		if err != nil {
			return err
		}
		p.token = token
		p.Send(p.PID(), watchTick{})
	case watchTick:
		p.seq++
		p.SendEvent(p.name, p.token, samplePayload(p.name, p.seq))
		p.SendAfter(p.PID(), watchTick{}, p.interval)
	}
	return nil
}

func (p *timerProducer) Terminate(reason error) { p.UnregisterEvent(p.name) }

// ── audit_trail: high rate, Notify=false (lower the window limit to see Suppressed) ──

func factoryAuditTrail() gen.ProcessBehavior { return &auditTrail{} }

type auditTrail struct {
	act.Actor
	token gen.Ref
	seq   int
}

func (p *auditTrail) Init(args ...any) error {
	p.SendAfter(p.PID(), msgRegister{}, watchRegisterDelay)
	return nil
}

func (p *auditTrail) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case msgRegister:
		token, err := p.RegisterEvent(evAudit, gen.EventOptions{Notify: false, Buffer: 50})
		if err != nil {
			return err
		}
		p.token = token
		p.Send(p.PID(), watchTick{})
	case watchTick:
		actions := []string{"login", "update", "delete", "export"}
		for i := 0; i < 5; i++ {
			p.seq++
			p.SendEvent(evAudit, p.token, AuditEntry{Seq: p.seq, Action: actions[p.seq%len(actions)], Actor: fmt.Sprintf("user%d", p.seq%9)})
		}
		p.SendAfter(p.PID(), watchTick{}, 25*time.Millisecond)
	}
	return nil
}

func (p *auditTrail) Terminate(reason error) { p.UnregisterEvent(evAudit) }

// ── deploy_status: registered for a deploy then unregistered, on a loop → "closed" ──

func factoryDeployStatus() gen.ProcessBehavior { return &deployStatus{} }

type deployStatus struct {
	act.Actor
	token      gen.Ref
	registered bool
	seq        int
}

func (p *deployStatus) Init(args ...any) error {
	p.SendAfter(p.PID(), msgRegister{}, watchRegisterDelay)
	return nil
}

func (p *deployStatus) register() error {
	token, err := p.RegisterEvent(evDeploy, gen.EventOptions{Notify: false, Buffer: 5})
	if err != nil {
		return err
	}
	p.token = token
	p.registered = true
	return nil
}

func (p *deployStatus) HandleMessage(from gen.PID, message any) error {
	switch message {
	case msgRegister{}:
		if err := p.register(); err != nil {
			return err
		}
		p.Send(p.PID(), "pub")
		p.SendAfter(p.PID(), "cycle", 20*time.Second)
	case "pub":
		if p.registered == false {
			break
		}
		p.seq++
		phases := []string{"build", "push", "rollout", "done"}
		p.SendEvent(evDeploy, p.token, DeployStatus{Release: fmt.Sprintf("v1.4.%d", p.seq), Phase: phases[p.seq%len(phases)], Progress: (p.seq * 10) % 100})
		p.SendAfter(p.PID(), "pub", 2*time.Second)
	case "cycle":
		if p.registered {
			p.UnregisterEvent(evDeploy) // deploy finished → event goes away
			p.registered = false
			p.SendAfter(p.PID(), "cycle", 8*time.Second)
			break
		}
		p.register()
		p.Send(p.PID(), "pub")
		p.SendAfter(p.PID(), "cycle", 20*time.Second)
	}
	return nil
}

func (p *deployStatus) Terminate(reason error) {
	if p.registered {
		p.UnregisterEvent(evDeploy)
	}
}

// ── chat member: joins/leaves chat_room so the watch banner flips live ──

func factoryChatMember() gen.ProcessBehavior { return &chatMember{} }

type chatMember struct {
	act.Actor
	target     gen.Event
	subscribed bool
}

func (s *chatMember) Init(args ...any) error {
	s.target = gen.Event{Name: evChat, Node: s.Node().Name()}
	s.SendAfter(s.PID(), "toggle", 3*time.Second)
	return nil
}

func (s *chatMember) HandleMessage(from gen.PID, message any) error {
	if message == "toggle" {
		if s.subscribed {
			s.DemonitorEvent(s.target)
			s.subscribed = false
		} else if _, err := s.MonitorEvent(s.target); err == nil {
			s.subscribed = true
		}
		s.SendAfter(s.PID(), "toggle", 12*time.Second)
	}
	return nil
}

func (s *chatMember) HandleEvent(message gen.MessageEvent) error { return nil }

func (s *chatMember) Terminate(reason error) {}
