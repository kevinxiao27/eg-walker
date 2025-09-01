package ol

type ID struct { // GUID
	agent string
	seq   int
}

func (id ID) Unpack() (string, int) {
	return id.agent, id.seq
}

type LV int // index for op log

type OpType string

const (
	Insert OpType = "ins"
	Delete OpType = "del"
)

type InnerOp[T any] struct {
	optype  OpType
	pos     int
	content T // Only meaningful for Insert/Retain
}

type Op[T any] struct {
	InnerOp[T]
	id      ID
	parents []LV
}

type RemoteVersion map[string]int // [agent] : last known sequence number

type OpLog[T any] struct {
	Ops      []Op[T]
	frontier []LV
	version  RemoteVersion
}
