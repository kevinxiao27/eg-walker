package eg

import (
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/kevinxiao27/eg-walker/util"
)

func expandLVToSet[T any](oplog OpLog[T], frontier []LV) mapset.Set[LV] {
	set := mapset.NewSet[LV]()
	toExpand := make([]LV, len(frontier))
	copy(toExpand, frontier)

	for len(toExpand) > 0 {
		lv := toExpand[len(toExpand)-1]
		toExpand = toExpand[:len(toExpand)-1]
		if set.Contains(lv) {
			continue
		}

		set.Add(lv)
		op := oplog.ops[lv]
		toExpand = append(toExpand, op.parents...)
	}

	return set

}

type DiffResult struct {
	aOnly []LV
	bOnly []LV
}

func retreat[T any](doc *CRDTDoc, oplog OpLog[T], opLV LV) {
	op := oplog.ops[opLV]
	var target LV
	if op.optype == Insert {
		target = opLV
	} else {
		target = (*doc.delTargets)[opLV]
	}

	item := doc.itemsByLV[target]
	item.curState-- // INS -> NYI -> D-0 -> ... -> D-N
}

func advance[T any](doc *CRDTDoc, oplog OpLog[T], opLV LV) {
	op := oplog.ops[opLV]
	var target LV
	if op.optype == Insert {
		target = opLV
	} else {
		target = (*doc.delTargets)[opLV]
	}

	item := doc.itemsByLV[target]
	item.curState++ // D_N -> ... -> D-0 -> NYI -> I
}

func findByCurrentPos(items []*CRDTItem, targetPos int) (idx int, endPos int) {
	curPos := 0
	endPos = 0
	idx = 0

	for curPos < targetPos {
		if idx >= len(items) {
			// If we've gone past the end of items, we can't find the target position
			break
		}

		item := items[idx]

		// skip CRDTs in NYI state, and deleted state
		// only care about inserted
		if item.curState == INSERTED {
			curPos++
		}
		if !item.deleted {
			endPos++
		}
		idx++
	}

	return idx, endPos
}

func findItemIdxAtLV(items []*CRDTItem, target LV) int {
	for i := range len(items) {
		if items[i].lv == target {
			return i
		}
	}

	return -1
}

func integrate[T any](oplog OpLog[T], doc *CRDTDoc, newItem *CRDTItem, idx int, endPos int, snapshot *[]T) {
	scanIdx := idx
	scanEndPos := endPos

	// if originLeft is null, we can pretend we're inserting to the right of -1
	left := scanIdx - 1
	right := util.Choose(newItem.originRight == LV(-1), len(*doc.items), findItemIdxAtLV(*doc.items, newItem.originRight))
	if right == -1 {
		right = len(*doc.items)
	}
	scanning := false

	for scanIdx < right {
		other := (*doc.items)[scanIdx]
		if other.curState != NOT_YET_INSERTED {
			break
		}

		// -1 is our stand in for unset
		oLeft := util.Choose(other.originLeft == LV(-1), -1, findItemIdxAtLV(*doc.items, other.originLeft))
		oRight := util.Choose(other.originRight == LV(-1), len(*doc.items), findItemIdxAtLV(*doc.items, other.originRight))

		// Compare with TypeScript logic
		newItemAgent := oplog.ops[newItem.lv].id.agent
		otherAgent := oplog.ops[other.lv].id.agent

		fmt.Printf(newItemAgent, otherAgent, newItemAgent < otherAgent)
		if oLeft < left ||
			(oLeft == left && oRight == right && newItemAgent < otherAgent) {
			break
		}

		if oLeft == left {
			scanning = oRight < right
		}

		if !other.deleted {
			scanEndPos++
		}
		scanIdx++

		if !scanning {
			idx = scanIdx
			endPos = scanEndPos
		}
	}

	(*doc.items) = append((*doc.items)[:idx], append([]*CRDTItem{newItem}, (*doc.items)[idx:]...)...)

	op := oplog.ops[newItem.lv]
	if op.optype != Insert {
		print("Something horrible has gone wrong and we cannot insert a delete\n")
	}
	*snapshot = append((*snapshot)[:endPos], append([]T{op.content}, (*snapshot)[endPos:]...)...)
}

func apply[T any](doc *CRDTDoc, oplog OpLog[T], snapshot *[]T, opLV LV) {
	op := oplog.ops[opLV]

	if op.optype == Delete {
		// need to do the following
		// 1. take doc and find item, mark isDeleted as true
		// 2. modify snapshot (actually remove the item)
		// pos may only be used when we have replayed all parent history

		idx, endPos := findByCurrentPos(*doc.items, op.pos)

		// scan forwards to find actual item
		for (*doc.items)[idx].curState != INSERTED {
			if !(*doc.items)[idx].deleted {
				endPos++
			}
			idx++
		}

		// This is the item to delete
		item := (*doc.items)[idx]
		if !item.deleted {
			item.deleted = true

			*snapshot = append((*snapshot)[:endPos], (*snapshot)[endPos+1:]...)

		}

		item.curState = 1
		(*doc.delTargets)[opLV] = item.lv

	} else {
		// INSERT
		idx, endPos := findByCurrentPos(*doc.items, op.pos)

		// item will definitely be in inserted state, findByCurrentPos
		// invariably terminates as a result of encountering an insertion
		var originLeft LV
		if idx == 0 {
			originLeft = -1
		} else {
			originLeft = (*doc.items)[idx-1].lv
		}

		//scan for originRight starting from idx
		originRight := LV(-1)
		for i := idx; i < len((*doc.items)); i++ {
			tmp := (*doc.items)[i]
			if tmp.curState != NOT_YET_INSERTED {
				originRight = tmp.lv
				break
			}
		}

		item := &CRDTItem{
			lv:          opLV,
			originLeft:  originLeft,
			originRight: originRight,
			deleted:     false,
			curState:    INSERTED,
		}

		doc.itemsByLV[opLV] = item

		integrate(oplog, doc, item, idx, endPos, snapshot)
	}

}

func diff[T any](oplog OpLog[T], a []LV, b []LV) DiffResult {
	//  versions a/b are refer to the set of the version and all of their parents
	// slow implementation... we grab all parents in order to compare. and also compare all of them
	aExpand := expandLVToSet(oplog, a)
	bExpand := expandLVToSet(oplog, b)

	return DiffResult{
		aOnly: mapset.Sorted(aExpand.Difference(bExpand)),
		bOnly: mapset.Sorted(bExpand.Difference(aExpand)),
	}
}

func Checkout[T any](oplog OpLog[T]) []T {
	items := []*CRDTItem{}
	cv := make([]LV, 0)
	dt := make([]LV, 0)
	doc := CRDTDoc{
		items:          &items,
		currentVersion: &cv,
		delTargets:     &dt,
		itemsByLV:      make(map[LV]*CRDTItem),
	}

	snapshot := []T{}

	for lv := 0; lv < len(oplog.ops); lv++ {
		op := oplog.ops[lv]

		diff := diff(oplog, *doc.currentVersion, op.parents)
		aOnly, bOnly := diff.aOnly, diff.bOnly

		// retreat things not included
		for _, i := range aOnly {
			retreat(&doc, oplog, i)
		}

		for _, i := range bOnly {
			advance(&doc, oplog, i)
		}

		// apply operation
		apply(&doc, oplog, &snapshot, LV(lv))
		cv := []LV{LV(lv)}
		doc.currentVersion = &cv // suspicious lil use after free but all good
	}

	return snapshot
}
