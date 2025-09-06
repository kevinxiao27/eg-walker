package eg

import (
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/kevinxiao27/eg-walker/util"
	"github.com/sanity-io/litter"
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

func retreat[T any](doc CRDTDoc, oplog OpLog[T], opLV LV) {
	op := oplog.ops[opLV]
	target := util.Choose((op.optype == Insert), opLV, doc.delTargets[opLV])

	item := doc.itemsByLV[target]
	litter.Dump(doc)
	item.curState-- // INS -> NYI -> D-0 -> ... -> D-N
}

func advance[T any](doc CRDTDoc, oplog OpLog[T], opLV LV) {
	op := oplog.ops[opLV]
	target := util.Choose((op.optype == Insert), opLV, doc.delTargets[opLV])

	item := doc.itemsByLV[target]
	item.curState++ // D_N -> ... -> D-0 -> NYI -> I
}

func findByCurrentPos(items []CRDTItem, targetPos int) (idx int, endPos int) {
	curPos := 0
	endPos = 0

	for idx := 0; curPos < targetPos; idx++ {
		if idx >= len(items) {
			print("Attempted to index out of bounds of array, illegal state encountered: %v", idx)
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
	}

	return idx, endPos
}

func findItemIdxAtLV(items []CRDTItem, target LV) int {
	for i := 0; i < len(items); i++ {
		if items[i].lv == target {
			return i
		}
	}

	print("something terrible has occured and we were unable to find target in CRDTItems")

	return -1
}

func integrate[T any](oplog OpLog[T], doc CRDTDoc, newItem CRDTItem, idx int, endPos int, snapshot *[]T) {
	scanIdx := idx
	scanEndPos := endPos

	// if originLeft is null, we can pretend we're inserting to the right of -1
	left := scanIdx - 1
	right := util.Choose(newItem.originRight == LV(-1), len(doc.items), findItemIdxAtLV(doc.items, newItem.originRight))
	scanning := false

	for i := left + 1; ; i++ {
		other := doc.items[scanIdx]
		if other.curState != NOT_YET_INSERTED {
			break
		}

		// -1 is our stand in for unset
		oLeft := util.Choose(other.originLeft == LV(-1), -1, findItemIdxAtLV(doc.items, other.originLeft))
		oRight := util.Choose(other.originRight == LV(-1), len(doc.items), findItemIdxAtLV(doc.items, other.originRight))

		// if our
		if oLeft < left ||
			(oLeft == left && oRight == right && oplog.ops[newItem.lv].id.agent < oplog.ops[other.lv].id.agent) {
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

	doc.items = append(doc.items[:idx], append([]CRDTItem{newItem}, doc.items[idx:]...)...)

	op := oplog.ops[newItem.lv]
	if op.optype != Insert {
		print("Something horrible has gone wrong and we cannot insert a delete")
	}
	*snapshot = append((*snapshot)[:endPos], append([]T{op.content}, (*snapshot)[endPos:]...)...)
}

func apply[T any](doc CRDTDoc, oplog OpLog[T], snapshot *[]T, opLV LV) {
	op := oplog.ops[opLV]

	if op.optype == Delete {
		// need to do the following
		// 1. take doc and find item, mark isDeleted as true
		// 2. modify snapshot (actually remove the item)
		// pos may only be used when we have replayed all parent history

		idx, endPos := findByCurrentPos(doc.items, op.pos)

		// scan forwards to find actual item
		for doc.items[idx].curState != INSERTED {
			if !doc.items[idx].deleted {
				endPos++
			}
			idx++
		}

		if !doc.items[idx].deleted {
			doc.items[idx].deleted = true
			*snapshot = append((*snapshot)[:endPos-1], (*snapshot)[endPos+1:]...)
		}

		doc.items[idx].curState = 1
		doc.delTargets[opLV] = doc.items[idx].lv
	} else {
		// INSERT
		idx, endPos := findByCurrentPos(doc.items, op.pos)

		// item will definitely be in inserted state, findByCurrentPos
		// invariably terminates as a result of encountering an insertion
		var originLeft LV
		if idx == 0 {
			originLeft = -1
		} else {
			originLeft = doc.items[idx-1].lv
		}

		//scan
		originRight := LV(-1)
		for i := 0; i < len(doc.items); i++ {
			tmp := doc.items[i]
			if tmp.curState != NOT_YET_INSERTED {
				originRight = tmp.lv
				break
			}

		}

		item := CRDTItem{
			lv:          opLV,
			originLeft:  originLeft,
			originRight: originRight,
			deleted:     false,
			curState:    NOT_YET_INSERTED,
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
		aOnly: aExpand.Difference(bExpand).ToSlice(),
		bOnly: bExpand.Difference(aExpand).ToSlice(),
	}
}

func Checkout[T any](oplog OpLog[T]) []T {
	doc := CRDTDoc{
		items:          []CRDTItem{},
		currentVersion: []LV{},
		delTargets:     []LV{},
		itemsByLV:      []CRDTItem{},
	}

	snapshot := []T{}

	for lv := 0; lv < len(oplog.ops); lv++ {
		op := oplog.ops[lv]

		diff := diff(oplog, doc.currentVersion, op.parents)
		aOnly, bOnly := diff.aOnly, diff.bOnly

		// retreat things not included
		for _, i := range aOnly {
			fmt.Printf("retreat: %v\n", i)
			retreat(doc, oplog, i)
		}

		// refers to union difference of toExpand and currentVersion
		litter.Dump(doc)

		for _, i := range bOnly {
			fmt.Printf("advance: %v\n", i)
			advance(doc, oplog, i)
		}

		litter.Dump(doc)

		// apply operation
		fmt.Printf("apply: %v\n", lv) // add to []items
		apply(doc, oplog, &snapshot, LV(lv))
		doc.currentVersion = []LV{LV(lv)}
	}

	return snapshot
}
