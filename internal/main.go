package main

import (
	"fmt"

	"github.com/kevinxiao27/eg-walker/internal/types"
)

func main() {
	oplog := types.NewOpLog[rune]()
	types.LocalInsert(&oplog, "kev", 0, []rune{'H', 'e', 'l', 'l', 'o'})
	fmt.Printf("%v", oplog.Ops)
}
