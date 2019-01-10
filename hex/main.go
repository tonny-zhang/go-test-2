package main

import (
	"encoding/hex"
	"fmt"
	"log"
)

func main() {

	content := []byte("Go is an open source programming language.")

	fmt.Printf("%s", hex.Dump(content))

	src := []byte(hex.Dump(content))

	dst := make([]byte, hex.DecodedLen(len(src)))
	n, err := hex.Decode(dst, src)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", dst[:n])

}
