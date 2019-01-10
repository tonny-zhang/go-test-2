package req

import (
	"fmt"
	"log"
	"net/http"
)

// Req 请求
// test
func Req(url string, ch chan int) {
	res, err := http.Get(url)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("statuscode = %d\n", res.StatusCode)
	// _, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Printf("%s\n", robots)
	fmt.Printf("url [%s] loaded\n", url)
	ch <- 1
}
