package latency

// MessagePing is the async message sent from sender to remote worker pool
type MessagePing struct {
	Seq int
}

// messageBurst is an internal trigger for the sender to fire a burst
type messageBurst struct{}
