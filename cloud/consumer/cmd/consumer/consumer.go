package main

import (
	"github.com/ergo-services/ergo/apps/cloud"
	"github.com/ergo-services/ergo/etf"
	"github.com/ergo-services/ergo/gen"
	"github.com/ergo-services/ergo/node"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"

	"github.com/ergo-services/examples/cloud/consumer"
)

// consumerServer implementation of Server
type consumerServer struct {
	gen.Server
	log *logrus.Entry
}

type messageSend struct {
	to gen.ProcessID
}

type consumerServerState struct {
	thisNode node.Node
}

func (cs *consumerServer) Init(process *gen.ServerProcess, args ...etf.Term) error {
	cs.log = log.WithFields(log.Fields{
		// Uncomment lines below to extend log format with the node name and process id (etf.Pid)
		//	process.NodeName(): process.Self(),
	})

	// Keep the node.Node interface to get the connected peers list.
	// Use prosess.State for that.
	thisNode := process.Env(node.EnvKeyNode).(node.Node)
	state := &consumerServerState{
		thisNode: thisNode,
	}
	process.State = state

	// Registering the message type allows us not to think about marshaling/unmarshaling
	// from/to golang native types. Ergo will do it automatically.
	// Use 'Strict' mode - it makes the node drop connection if the message pretended
	// to be this type couldn't be decoded into the variable of this type.
	opts := etf.RegisterTypeOptions{
		Strict: true,
	}
	if name, err := etf.RegisterType(consumer.Message{}, opts); err == nil {
		cs.log.Infof("registered type: %q", name)
	} else {
		return err
	}

	// Enable monitoring the cloud connection state
	if err := process.MonitorEvent(cloud.EventCloud); err != nil {
		cs.log.Errorf("can not monitor event cloud.EventCloud:", err)
		return err
	}

	// Enable monitoring network events to receive updates on new connection or disconnection
	if err := process.MonitorEvent(node.EventNetwork); err != nil {
		cs.log.Errorf("can not monitor event node.EventNetwork:", err)
		return err
	}

	cs.log.Infof("process started with PID: %s", process.Self())
	return nil
}

func (cs *consumerServer) HandleInfo(process *gen.ServerProcess, message etf.Term) gen.ServerStatus {
	state := process.State.(*consumerServerState)
	switch m := message.(type) {
	case node.MessageEventNetwork:
		cs.log.Infof("network event: [node %q] online: %v", m.PeerName, m.Online)
		cs.log.Infof("connected nodes %v", state.thisNode.Nodes())
		cs.log.Infof("connected nodes via cloud ☁️  %v", state.thisNode.NodesIndirect())

	case cloud.MessageEventCloud:
		cs.log.Infof("cloud connection: [cluster %q via %q] online: %v", m.Cluster, m.Proxy, m.Online)

	case consumer.Message:
		cs.log.Infof("got message 📬 from %s (PID: %s)", m.From.Node, m.From)

	default:
		cs.log.Errorf("unhandled message: %#v\n", m)
	}
	return gen.ServerStatusOK
}
