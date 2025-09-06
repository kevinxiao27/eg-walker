package main

import (
	"github.com/kevinxiao27/eg-walker/eg"
	"github.com/sanity-io/litter"
)

func main() {
	litter.Config.HidePrivateFields = false
	oplog1 := eg.NewOpLog[rune]()
	oplog2 := eg.NewOpLog[rune]()
	eg.LocalInsert(&oplog1, "kev", 0, []rune("hi"))
	eg.LocalInsert(&oplog2, "kez", 0, []rune("yoooo"))

	eg.MergeInto(&oplog1, &oplog2)
	eg.MergeInto(&oplog2, &oplog1)

	eg.Checkout(oplog1)
	litter.Dump(oplog1)

	result := eg.Checkout(oplog1)
	litter.Dump(result)
}
