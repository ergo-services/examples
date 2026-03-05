package messaging

import "ergo.services/ergo/net/edf"

type OrderSide string

type OrderTag struct {
	Key   string
	Value string
}

type TestOrder struct {
	ID       string
	Exchange string
	Side     OrderSide
	Price    float64
	Amount   *float64
	StopLoss *float64
	Tags     []OrderTag
	Fills    []*OrderTag
	Metadata map[string]string
	Notes    *string
	Margin   bool
}

func init() {
	edf.RegisterTypeOf(MessagePayload{})
	edf.RegisterTypeOf(OrderSide(""))
	edf.RegisterTypeOf(OrderTag{})
	edf.RegisterTypeOf(TestOrder{})
}

// MessagePayload is the async message sent from sender to remote worker pool
type MessagePayload struct {
	Data string
}

// messageBurst is an internal trigger for the sender to fire a burst
type messageBurst struct{}
