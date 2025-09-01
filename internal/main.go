package main

import (
	"fmt"

	"github.com/kevinxiao27/eg-walker/internal/ol"
)

func main() {
	oplog := ol.NewOpLog[rune]()
	ol.LocalInsert(&oplog, "kev", 0, []rune{'H', 'e', 'l', 'l', 'o'})
	fmt.Printf("%+v\n", oplog)
}
