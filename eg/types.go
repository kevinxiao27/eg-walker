package eg

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
	ops      []Op[T]
	frontier []LV
	version  RemoteVersion
}

const (
	NOT_YET_INSERTED int = -1
	INSERTED         int = 0
)

type CRDTItem struct {
	lv          LV
	originLeft  LV // -1 as not set value
	originRight LV
	deleted     bool
	curState    int
}

type CRDTDoc struct {
	items          []CRDTItem
	currentVersion []LV
	delTargets     []LV       // LV of delete OP
	itemsByLV      []CRDTItem // Map from LV => CRDTItem
}
