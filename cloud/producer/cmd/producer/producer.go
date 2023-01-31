package main

import (
	"time"

	"github.com/ergo-services/ergo/apps/cloud"
	"github.com/ergo-services/ergo/etf"
	"github.com/ergo-services/ergo/gen"
	"github.com/ergo-services/ergo/node"
	"github.com/ergo-services/examples/cloud/consumer"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// producerServer implementation of Server
type producerServer struct {
	gen.Server
	log *logrus.Entry
}

type messageSendLocal struct{}
type messageSendCloud struct{}

type producerServerState struct {
	thisNode node.Node
	toLocal  gen.ProcessID
	toCloud  gen.ProcessID
}

func (ps *producerServer) Init(process *gen.ServerProcess, args ...etf.Term) error {
	ps.log = log.WithFields(log.Fields{
		// Uncomment line bellow to extend log format with the node name and process id (etf.Pid)
		// process.NodeName(): process.Self(),
	})

	// Keep the node.Node interface to get the connected peers list.
	// Use prosess.State for that.
	thisNode := process.Env(node.EnvKeyNode).(node.Node)
	state := &producerServerState{
		thisNode: thisNode,
	}
	process.State = state

	state.toLocal = gen.ProcessID{Name: "demo", Node: OptionTo + "@localhost"}
	state.toCloud = gen.ProcessID{Name: "demo", Node: OptionTo + "@" + OptionCloudClusterName}

	// Registering the message type allows us not to think about marshaling/unmarshaling
	// from/to golang native types. Ergo will do it automatically.
	// Use 'Strict' mode - it makes the node drop connection if the message pretended
	// to be this type couldn't be decoded into the variable of this type.
	opts := etf.RegisterTypeOptions{
		Strict: true,
	}
	if name, err := etf.RegisterType(consumer.Message{}, opts); err == nil {
		ps.log.Infof("registered type: %q", name)
	} else {
		return err
	}

	// Enable monitoring the cloud connection events
	if err := process.MonitorEvent(cloud.EventCloud); err != nil {
		ps.log.Errorf("can not monitor event cloud.EventCloud:", err)
		return err
	}

	// Enable monitoring network events to receive updates on new connection or disconnection
	if err := process.MonitorEvent(node.EventNetwork); err != nil {
		ps.log.Errorf("can not monitor event node.EventNetwork:", err)
		return err
	}

	ps.log.Infof("process started with PID: %s", process.Self())
	return nil
}

func (ps *producerServer) HandleInfo(process *gen.ServerProcess, message etf.Term) gen.ServerStatus {
	// get the process' state
	state := process.State.(*producerServerState)
	switch m := message.(type) {
	case messageSendLocal:
		ps.log.Infof("sending message ✉️  to: %s", state.toLocal)
		// send message to the consumer
		if err := process.Send(state.toLocal, consumer.Message{From: process.Self()}); err != nil {
			ps.log.Errorf("can not send message: %s", err)
		}
		// schedule sending the next message to the consumer via cloud connection
		process.SendAfter(process.Self(), messageSendCloud{}, time.Second)

	case messageSendCloud:
		ps.log.Infof("sending message ✉️  to: %s (via cloud ☁️ ) ", state.toCloud)
		// send message to the consumer
		if err := process.Send(state.toCloud, consumer.Message{From: process.Self()}); err != nil {
			ps.log.Errorf("can not send message: %s", err)
		}
		// schedule sending the next message to the consumer via local connection
		process.SendAfter(process.Self(), messageSendLocal{}, time.Second)

	case node.MessageEventNetwork:
		// as long as we monitor event node.EventNetwork, we are receiving these messages
		ps.log.Infof("network event: [node %q] online: %v", m.PeerName, m.Online)
		ps.log.Infof("connected nodes %v", state.thisNode.Nodes())
		ps.log.Infof("connected nodes via cloud ☁️  %v", state.thisNode.NodesIndirect())

	case cloud.MessageEventCloud:
		ps.log.Infof("cloud connection: [cluster %q via %q] online: %v", m.Cluster, m.Proxy, m.Online)
		if m.Online == true {
			ps.log.Infof("start sending a message...")
			process.Send(process.Self(), messageSendLocal{})
		}

	default:
		ps.log.Errorf("unhandled message: %#v", m)
	}
	return gen.ServerStatusOK
}
