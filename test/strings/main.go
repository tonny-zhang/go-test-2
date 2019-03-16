package main

import (
	"fmt"
	"strings"
)

func main() {
	pathOld := "a/b/c"
	fmt.Println(strings.Join(strings.Split(pathOld, "\\"), "/"))
	pathNew := strings.Replace(pathOld, "\\", "/", 0)
	fmt.Println(pathNew)
}
