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
	eg.LocalInsert(&oplog1, "a", 0, []rune("hi"))
	eg.LocalInsert(&oplog2, "z", 0, []rune("yoooo"))

	eg.MergeInto(&oplog1, &oplog2)
	eg.MergeInto(&oplog2, &oplog1)

	result1 := eg.Checkout(oplog1)
	fmt.Printf("Result: %v → '%s'\n", result1, string(result1))

	result2 := eg.Checkout(oplog2)
	fmt.Printf("Result: %v → '%s'\n", result2, string(result2))

}
