package main

import (
	"fmt"

	"github.com/kevinxiao27/eg-walker/eg"
	"github.com/sanity-io/litter"
)

func main() {
	litter.Config.HidePrivateFields = false
	oplog1 := eg.NewOpLog[rune]()
	oplog2 := eg.NewOpLog[rune]()
	oplog3 := eg.NewOpLog[rune]()
	eg.LocalInsert(&oplog1, "a", 0, []rune("I'm like hey"))
	eg.LocalInsert(&oplog3, "z", 0, []rune(" hello"))
	eg.LocalInsert(&oplog2, "c", 0, []rune(" wassup"))

	eg.MergeInto(&oplog1, &oplog2)
	eg.MergeInto(&oplog2, &oplog1)
	eg.MergeInto(&oplog2, &oplog3)
	eg.MergeInto(&oplog3, &oplog2)
	eg.MergeInto(&oplog1, &oplog2)

	result1 := eg.Checkout(oplog1)
	fmt.Printf("Result: %v → '%s'\n", result1, string(result1))

	result2 := eg.Checkout(oplog2)
	fmt.Printf("Result: %v → '%s'\n", result2, string(result2))

	result3 := eg.Checkout(oplog3)
	fmt.Printf("Result: %v → '%s'\n", result3, string(result3))
}
