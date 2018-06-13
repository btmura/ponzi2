package main

import (
	"flag"
	"fmt"
)

var symbol = flag.String("s", "SPY", `Stock symbol like "SPY"`)

func main() {
	flag.Parse()
	fmt.Printf("Symbol: %s", *symbol)
}
