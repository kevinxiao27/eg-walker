package types

type ID struct { // GUID
	agent string
	seq   int
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
	ops      []Op[T]
	frontier []LV
	version  RemoteVersion
}

func NewOpLog[T any]() OpLog[T] {
	return OpLog[T]{
		ops:      []Op[T]{},
		frontier: []LV{},
		version:  make(map[string]int),
	}
}

func appendLocalOp[T any](oplog OpLog[T], agent string, op InnerOp[T]) {
	seq := oplog.version[agent] + 1
	lv := len(oplog.ops)

	oplog.ops = append(oplog.ops, Op[T]{
		InnerOp: op,
		id:      ID{agent, seq},
		parents: oplog.frontier,
	})

	oplog.frontier = []LV{LV(lv)}
	oplog.version[agent] = seq
}

func LocalInsert[T any](oplog OpLog[T], agent string, pos int, content T) {
	appendLocalOp(oplog, agent, InnerOp[T]{optype: Insert, pos: pos, content: content})
}

func LocalDelete[T any](oplog OpLog[T], agent string, pos int) {
	appendLocalOp(oplog, agent, InnerOp[T]{optype: Delete, pos: pos})
}
