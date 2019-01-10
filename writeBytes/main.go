package main

import (
	"encoding/binary"
	"fmt"
)

func main() {
	data := make([]byte, 20)
	fmt.Println(data)
	binary.LittleEndian.PutUint32(data[2:], 16520)

	fmt.Println(data)
}
