package latency

import "ergo.services/ergo/net/edf"

func init() {
	if err := edf.RegisterTypeOf(MessagePing{}); err != nil {
		panic(err)
	}
}

// MessagePing is the async message sent from sender to remote worker pool
type MessagePing struct {
	Seq int
}

// messageBurst is an internal trigger for the sender to fire a burst
type messageBurst struct{}
