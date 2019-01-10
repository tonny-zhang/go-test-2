package main

import (
	"fmt"
	"test/test/http/req"
)

func main() {
	ch1 := make(chan int)
	ch2 := make(chan int)
	go req.Req("https://www.baidu.com", ch1)

	go req.Req("https://www.baidu.com/?key=", ch2)

	fmt.Println("after req")

	<-ch1
	<-ch2

	fmt.Println("after req1")
}
