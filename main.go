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

	if len(result1) == len(result2) {
		fmt.Println("Lengths match")
	} else {
		fmt.Println("Lengths differ")
	}

	fmt.Println("Character positions:")
	for i := 0; i < len(result1); i++ {
		if result1[i] != result2[i] {
			fmt.Printf("Position %d differs: oplog1=%d, oplog2=%d\n", i, result1[i], result2[i])
		}
	}
}
