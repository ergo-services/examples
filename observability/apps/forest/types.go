package forest

import "ergo.services/ergo/gen"

type MessageCompute struct {
	Seq     uint64
	Payload []byte
}

type MessageIngest struct {
	Stream gen.Atom
	Value  int64
}

type MessageJob struct {
	ID uint64
}
