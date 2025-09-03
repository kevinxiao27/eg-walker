package eg

import (
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
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

func Diff[T any](oplog OpLog[T], a []LV, b []LV) DiffResult {
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
	currentVersion := []LV{}

	for lv := 0; lv < len(oplog.Ops); lv++ {
		op := oplog.Ops[lv]

		diff := Diff(oplog, currentVersion, op.parents)
		aOnly, bOnly := diff.aOnly, diff.bOnly

		// retreat things not included
		for i := range aOnly {
			fmt.Printf("retreat: %v\n", i)
			// TODO
		}

		// refers to union difference of toExpand and currentVersion

		for i := range bOnly {
			fmt.Printf("advance: %v\n", i)
		}

		// apply operation
		fmt.Printf("apply: %v\n", lv)
		currentVersion = []LV{LV(lv)}
	}

	return []T{}
}
