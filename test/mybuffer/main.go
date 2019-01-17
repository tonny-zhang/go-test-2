package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"test/test/mybuffer/code"
)

const typeObject byte = 100
const typeString byte = 7
const typeArray byte = 101
const typeInt byte = 1

func _getTypeStr(vType byte) string {
	switch vType {
	case typeInt:
		return "int"
	}
	return ""
}
func main1() {
	filename := "./conf/test.json"
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return
	}

	conf := make(map[string]interface{})

	err = json.Unmarshal(data, &conf)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(conf)
	var content = "package code\n\n"
	// name := conf["name"]
	typeVal := byte(conf["type"].(float64))
	propVal := conf["prop"]

	if typeVal == typeObject {
		content += "type Code1 struct {\n"

		propVal := propVal.([]interface{})
		for _, prop := range propVal {
			pv := prop.(map[string]interface{})
			tVal := byte(pv["type"].(float64))
			name := pv["name"].(string)
			switch tVal {
			case typeString:
				content += "\t" + name + " string\n"
			case typeArray:
				content += "\t" + name + " []" + _getTypeStr(byte(pv["prop"].(map[string]interface{})["type"].(float64))) + "\n"
			}
		}
		content += "}"
	}

	fmt.Println(content)

	f, err := os.Create("./code/Code1.go")
	defer f.Close()
	if err == nil {
		f.WriteString(content)
	} else {
		fmt.Println(err)
	}
}

func main() {
	a := code.Code1{}
	a.SetIds([]int32{1, 2, 3})
	a.SetIsCaptain("我们")
	fmt.Println(a)
	b := a.GetBytes()
	fmt.Println(a)
	fmt.Println(a.ParseFrom(b))
}
