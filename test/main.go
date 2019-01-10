package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

const (
	configFileUrl string = "./config.txt"
)

var flagMap = map[string]string{
	"line": "style=\"width:100%; margin:0 auto;\"",
	"img3": "style=\"display:block; margin:10px; width:33%; height:280px; float:left;",
	"img2": "style=\"display:block; margin:10px; width:49%; height:300px; float:left;",
	"img1": "style=\"display:block; margin:10px; width:900px; height:400px;",
}

var genHTML string = ""

func main() {

	file, err := os.Open(configFileUrl)

	if err != nil {
		fmt.Println(err)
	}

	defer file.Close()

	bufR := bufio.NewReader(file)

	for {
		_line, _, c := bufR.ReadLine()

		if c == io.EOF {
			break
		}

		_domName, _domFlag, _otherFlag, _href := getDom(string(_line))

		genHTML += genLine(_domName, _domFlag, _otherFlag, _href)
	}

	write(genHTML)
}

func getDom(line string) (string, string, string, string) {
	_list := strings.Split(line, " ")

	var needAdd = 4 - len(_list)

	for i := 0; i < needAdd; i++ {
		_list = append(_list, "")
	}

	return _list[0], _list[1], _list[2], _list[3]
}

func genLine(DomName string, flag string, bg string, href string) string {

	_flag := flagMap[flag]

	if strings.Index(bg, "bg") > -1 {

		flags := strings.Split(bg, "~")

		bg = "background:url(" + flags[1] + ")\""
	}

	if strings.Index(href, "href") > -1 {
		flags := strings.Split(href, "~")

		href = "href=" + "\"" + flags[1]
	}

	if strings.Index(DomName, "div") == -1 {
		href += "\""
	}
	return "<" + DomName + " " + _flag + bg + href + ">"
}

func write(_this string) {
	err := ioutil.WriteFile("./tb.html", []byte(_this), 0666)

	if err != nil {
		panic(err)
	}
}
