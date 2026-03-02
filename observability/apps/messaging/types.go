package messaging

import "ergo.services/ergo/net/edf"

func init() {
	if err := edf.RegisterTypeOf(MessagePayload{}); err != nil {
		panic(err)
	}
}

// MessagePayload is the async message sent from sender to remote worker pool
type MessagePayload struct {
	Data string
}

// messageBurst is an internal trigger for the sender to fire a burst
type messageBurst struct{}
