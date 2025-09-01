package main

import (
	"fmt"

	"github.com/kevinxiao27/eg-walker/internal/ol"
)

func main() {
	oplog1 := ol.NewOpLog[rune]()
	oplog2 := ol.NewOpLog[rune]()
	ol.LocalInsert(&oplog1, "kev", 0, []rune{'a'})
	ol.LocalInsert(&oplog2, "kez", 0, []rune{'b'})

	ol.MergeInto(&oplog1, &oplog2)
	ol.MergeInto(&oplog2, &oplog1)
	fmt.Printf("%+v\n", oplog1)
	fmt.Printf("%+v\n", oplog2)

}
