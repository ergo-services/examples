package tracing

import "ergo.services/ergo/net/edf"

func init() {
	edf.RegisterTypeOf(MessagePing{})
	edf.RegisterTypeOf(MessagePong{})
	edf.RegisterTypeOf(MessageNotify{})
	edf.RegisterTypeOf(MessageStatus{})
	edf.RegisterTypeOf(PingRequest{})
	edf.RegisterTypeOf(PongResponse{})
	edf.RegisterTypeOf(ValidateRequest{})
	edf.RegisterTypeOf(ValidateResponse{})
	edf.RegisterTypeOf(MessageForward{})
}

// MessagePing async fire-and-forget message
type MessagePing struct {
	Seq  int
	From string
}

// MessagePong async response to ping
type MessagePong struct {
	Seq  int
	From string
}

// MessageNotify async notification
type MessageNotify struct {
	Kind    string
	Payload string
}

// MessageStatus async status update
type MessageStatus struct {
	Service string
	Status  string
}

// PingRequest sync call request
type PingRequest struct {
	Payload string
}

// PongResponse sync call response
type PongResponse struct {
	Payload string
	Node    string
}

// ValidateRequest sync validation call
type ValidateRequest struct {
	OrderID string
	Amount  int
}

// ValidateResponse sync validation response
type ValidateResponse struct {
	Valid  bool
	Reason string
}

// MessageForward request to be forwarded to another node
type MessageForward struct {
	OriginalSender string
	Hops           []string
	Payload        string
}

// messageTick internal timer
type messageTick struct{}
