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
	Ops      []Op[T]
	frontier []LV
	version  RemoteVersion
}

func NewOpLog[T any]() OpLog[T] {
	return OpLog[T]{
		Ops:      []Op[T]{},
		frontier: []LV{},
		version:  make(map[string]int),
	}
}

func appendLocalOp[T any](oplog *OpLog[T], agent string, op InnerOp[T]) {
	seq := oplog.version[agent] + 1
	lv := len(oplog.Ops)

	oplog.Ops = append(oplog.Ops, Op[T]{
		InnerOp: op,
		id:      ID{agent, seq},
		parents: oplog.frontier,
	})

	oplog.frontier = []LV{LV(lv)}
	oplog.version[agent] = seq
}

func LocalInsert[T any](oplog *OpLog[T], agent string, pos int, content any) {
	if str, ok := content.(string); ok {
		if stringOpLog, ok := any(oplog).(*OpLog[string]); ok { // honestly disgusting but I can't think of something better
			for _, c := range str {
				appendLocalOp[string](stringOpLog, agent, InnerOp[string]{
					optype:  Insert,
					pos:     pos,
					content: string(c),
				})
				pos++
			}
			return
		}
	}

	if slice, ok := content.([]T); ok {
		for _, c := range slice {
			appendLocalOp(oplog, agent, InnerOp[T]{
				optype:  Insert,
				pos:     pos,
				content: c,
			})
			pos++
		}
		return
	}
}

func LocalDelete[T any](oplog *OpLog[T], agent string, pos int, delLen int) {
	for i := delLen; i > 0; i-- {
		appendLocalOp(oplog, agent, InnerOp[T]{optype: Delete, pos: pos})
		// pos doesn't need to be modified as proceeding characters will elide
	}
}
