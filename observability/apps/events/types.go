package events

import "ergo.services/ergo/net/edf"

func init() {
	if err := edf.RegisterTypeOf(MessageEventData{}); err != nil {
		panic(err)
	}
}

// MessageEventData is the payload published through events
type MessageEventData struct {
	Seq int
}

// messagePublish is an internal trigger for the publisher to send an event
type messagePublish struct{}

// messageStartPublishers triggers publisher startup in the SOFO supervisor
type messageStartPublishers struct{}

// messageStartSubscribers triggers subscriber startup in the SOFO supervisor
type messageStartSubscribers struct{}
