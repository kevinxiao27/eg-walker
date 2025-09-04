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
		op := oplog.Ops[lv]
		toExpand = append(toExpand, op.parents...)
	}

	return set

}

type DiffResult struct {
	aOnly []LV
	bOnly []LV
}

func retreat[T any](doc CRDTDoc, oplog OpLog[T], opLV LV) {
	op := oplog.Ops[opLV]
	target := util.Choose((op.optype == Insert), opLV, doc.delTargets[opLV])

	item := doc.itemsByLV[target]
	item.curState-- // INS -> NYI -> D-0 -> ... -> D-N
}

func advance[T any](doc CRDTDoc, oplog OpLog[T], opLV LV) {
	op := oplog.Ops[opLV]
	target := util.Choose((op.optype == Insert), opLV, doc.delTargets[opLV])

	item := doc.itemsByLV[target]
	item.curState++ // D_N -> ... -> D-0 -> NYI -> I
}

func apply[T any](doc CRDTDoc, oplog OpLog[T], snapshot []T, opLV LV) {
	op := oplog.Ops[opLV]

	if op.optype == Insert {
		// insert
	} else {
		// delete
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

	for lv := 0; lv < len(oplog.Ops); lv++ {
		op := oplog.Ops[lv]

		diff := diff(oplog, doc.currentVersion, op.parents)
		aOnly, bOnly := diff.aOnly, diff.bOnly

		// retreat things not included
		for _, i := range aOnly {
			fmt.Printf("retreat: %v\n", i)
			retreat(doc, oplog, i)
		}

		// refers to union difference of toExpand and currentVersion

		for _, i := range bOnly {
			fmt.Printf("advance: %v\n", i)
			advance(doc, oplog, i)
		}

		// apply operation
		fmt.Printf("apply: %v\n", lv) // add to []items
		doc.currentVersion = []LV{LV(lv)}
	}

	return []T{}
}
