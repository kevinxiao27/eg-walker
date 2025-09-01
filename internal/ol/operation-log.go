package ol

import (
	"fmt"
	"sort"

	"github.com/kevinxiao27/eg-walker/internal/util"
)

type ID struct { // GUID
	agent string
	seq   int
}

func IdEq(a ID, b ID) bool {
	return a.agent == b.agent && a.seq == b.seq
}

func (id ID) Unpack() (string, int) {
	return id.agent, id.seq
}

type LV int // index for op log

func sortLV(frontier []LV) []LV {
	sort.Slice(frontier, func(i, j int) bool {
		return frontier[i] > frontier[j]
	})
	return frontier
}

func advanceFrontier(frontier []LV, lv LV, parents []LV) []LV {
	f := util.Filter(frontier, func(lv LV) bool {
		return !util.Reduce(parents, func(lvInner LV, exists bool) bool {
			return lv == lvInner || exists
		}, false)
	})

	f = append(f, lv)
	sortLV(f)
	return f
}

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

func mapIDtoLV[T any](oplog *OpLog[T], id ID) (LV, error) {
	// optimization uses B-tree

	for i, op := range oplog.Ops {
		if IdEq(op.id, id) {
			return LV(i), nil
		}
	}

	return LV(-1), fmt.Errorf("could not find id in oplog")
}

func PushRemoteOp[T any](oplog *OpLog[T], op Op[T], parentIds []ID) {
	agent, seq := op.id.Unpack()
	lastKnownSeq := -1
	if v, ok := oplog.version[agent]; ok {
		lastKnownSeq = v
	}

	if lastKnownSeq >= seq { // already included
		return
	}

	lv := LV(len(oplog.Ops))

	funcParentsToLV := func(id ID) (LV, error) {
		return mapIDtoLV(oplog, id)
	}
	parents := sortLV(util.MapN[ID, LV](parentIds, funcParentsToLV))

	oplog.Ops = append(oplog.Ops, Op[T]{InnerOp: op.InnerOp, id: op.id, parents: parents})
	oplog.frontier = advanceFrontier(oplog.frontier, lv, parents)

	if lastKnownSeq+1 != seq {
		return
	}
	oplog.version[agent] = seq // assumes that seq = lastKnownSeq + 1, b tree implementation would likely not need this invariant

}

func MergeInto[T any](dest OpLog[T], src OpLog[T]) {
	// in real network call we would have to make some network calls
	// 1. find local seq -> 2. request remote ->
	// 3. remote returns all new changes since version -> 4. take events and merge

	// for _, op := range src.Ops {

	// }
	return
}
