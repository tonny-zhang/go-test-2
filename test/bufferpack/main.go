package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"test/test/bufferpack/buffer"
)

func create1() {
	val := 1
	var confJSON = `{"type":1, "prop": [{"type": 2, "name": "name"}]}`
	var conf interface{}
	err := json.Unmarshal([]byte(confJSON), &conf)
	if err != nil {
		log.Fatal(err)
	}

	confObj := conf.(map[string]interface{})

	for key, val := range confObj {
		fmt.Printf("%s is %T\n", key, val)
	}

	buf := new(bytes.Buffer)
	keyType := int(confObj["type"].(float64))
	fmt.Println(keyType)
	if keyType == 1 {
		binary.Write(buf, binary.LittleEndian, val)
		fmt.Printf("write int %d\n", val)
	}

	fmt.Printf("buf result = %d\n", binary.Size(buf))
	fmt.Println(buf.String())

}

func testCreateByte() {
	var confJSON = `{"type":6}`
	var conf interface{}
	err := json.Unmarshal([]byte(confJSON), &conf)
	if err != nil {
		log.Fatal(err)
	}

	var val byte = 1
	result, err := buffer.Create(conf, val)

	fmt.Printf("write byte error = %T, result = %x\n", err, result)
}
func testCreateInt16() {
	var confJSON = `{"type":1, "prop": [{"type": 2, "name": "name"}]}`
	var conf interface{}
	err := json.Unmarshal([]byte(confJSON), &conf)
	if err != nil {
		log.Fatal(err)
	}

	var int16Val int16 = 1
	result, err := buffer.Create(conf, int16Val)

	fmt.Printf("write int16 err = %v, result = %x\n", err, result)
}
func testCreateInt() {
	var confJSON = `{"type": 0}`
	var conf interface{}
	err := json.Unmarshal([]byte(confJSON), &conf)
	if err != nil {
		log.Fatal(err)
	}

	var intVal int32 = 1
	result, err := buffer.Create(conf, intVal)

	fmt.Printf("write int err = %v, result = %x\n", err, result)
}
func testCreateIntUseInt16() {
	var confJSON = `{"type": 0}`
	var conf interface{}
	err := json.Unmarshal([]byte(confJSON), &conf)
	if err != nil {
		log.Fatal(err)
	}

	var intVal int16 = 1
	result, err := buffer.Create(conf, intVal)

	fmt.Printf("write int use int16 err = %v, result = %x\n", err, result)
}
func testCreateLong() {
	var confJSON = `{"type": 5}`
	var conf interface{}
	err := json.Unmarshal([]byte(confJSON), &conf)
	if err != nil {
		log.Fatal(err)
	}

	var intVal int64 = 100
	result, err := buffer.Create(conf, intVal)

	fmt.Printf("write long err = %v, result = %x\n", err, result)
}
func testCreateFloat() {
	var confJSON = `{"type": 2}`
	var conf interface{}
	err := json.Unmarshal([]byte(confJSON), &conf)
	if err != nil {
		log.Fatal(err)
	}

	var val float32 = 100
	result, err := buffer.Create(conf, val)

	fmt.Printf("write float err = %v, result = %x\n", err, result)
}

func testCreateDouble() {
	var confJSON = `{"type": 3}`
	var conf interface{}
	err := json.Unmarshal([]byte(confJSON), &conf)
	if err != nil {
		log.Fatal(err)
	}

	var val float64 = 100
	result, err := buffer.Create(conf, val)

	fmt.Printf("write double err = %v, result = %x\n", err, result)
}
func testCreateBool() {
	var confJSON = `{"type": 7}`
	var conf interface{}
	err := json.Unmarshal([]byte(confJSON), &conf)
	if err != nil {
		log.Fatal(err)
	}

	var val = true
	result, err := buffer.Create(conf, val)

	fmt.Printf("write bool err = %v, result = %x\n", err, result)
}

func testCreateString() {
	var confJSON = `{"type": 4}`
	var conf interface{}
	err := json.Unmarshal([]byte(confJSON), &conf)
	if err != nil {
		log.Fatal(err)
	}

	var val = "hello"
	result, err := buffer.Create(conf, val)

	fmt.Printf("write string err = %v, result = %x\n", err, result)
}

func testCreateObject() {
	var confJSON = `{"type": 100, "prop": [
		{"type": 4, "name": "name"}, 
		{"type": 1, "name": "age"},
		{"type": 2, "name": "height"},
		{"type": 100, "name": "method", "prop": [{
			"type": 4, "name": "name"
		}]}
	]}`
	var conf interface{}
	err := json.Unmarshal([]byte(confJSON), &conf)
	if err != nil {
		log.Fatal(err)
	}

	// type Data struct {
	// 	Name string
	// 	Age  int16
	// }
	// d := Data{"tonny", 10}
	// fmt.Println(d.Name)
	// sysConfig := reflect.ValueOf(&d).Elem()
	// fmt.Printf("tt = %T\n", sysConfig.FieldByName("name"))
	obj := map[string]interface{}{
		"name":   "tonny",
		"age":    int16(10),
		"height": float32(0.21),
		"method": map[string]interface{}{
			"name": "run",
		},
	}
	result, err := buffer.Create(conf, obj)

	fmt.Printf("write object err = %v, result = %x\n", err, result)
}

func testCreateArray() {
	var confJSON = `{"type": 101, "prop": {
		"type": 1
	}}`
	var conf interface{}
	err := json.Unmarshal([]byte(confJSON), &conf)
	if err != nil {
		log.Fatal(err)
	}

	obj := []int16{
		1, 2, 3, 4, 5, 6,
	}
	result, err := buffer.Create(conf, obj)

	fmt.Printf("write array err = %v, result = %x\n", err, result)
}

func testCreateArrayObject() {
	var confJSON = `{"type": 102, "prop": [{
		"type": 0,
		"name": "id"
	}, {
		"type": 4,
		"name": "name"
	}]}`
	var conf interface{}
	err := json.Unmarshal([]byte(confJSON), &conf)
	if err != nil {
		log.Fatal(err)
	}

	obj := []map[string]interface{}{
		map[string]interface{}{
			"id":   1,
			"name": "one",
		},
		map[string]interface{}{
			"id":   2,
			"name": "two",
		},
	}
	result, err := buffer.Create(conf, obj)

	fmt.Printf("write arrayObject err = %v, result = %x\n", err, result)
}
func main() {

	testCreateInt16()

	testCreateInt()

	testCreateIntUseInt16()

	testCreateLong()

	testCreateFloat()
	testCreateDouble()
	testCreateByte()
	testCreateBool()
	testCreateString()
	testCreateObject()
	testCreateArray()

	testCreateArrayObject()
}
