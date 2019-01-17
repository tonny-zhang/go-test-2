package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Code1 struct {
	ids        []int32
	is_captain string
}

func (code *Code1) SetIds(ids []int32) {
	code.ids = ids
}
func (code Code1) GetIds() []int32 {
	return code.ids
}
func (code *Code1) SetIsCaptain(is_captain string) {
	code.is_captain = is_captain
}
func (code Code1) GetIsCaptain() string {
	return code.is_captain
}

// GetBytes a
func (code Code1) GetBytes() []byte {
	buf := new(bytes.Buffer)
	fmt.Println(len(code.ids))
	err := binary.Write(buf, binary.LittleEndian, int32(len(code.ids)))

	fmt.Println(err)
	for _, v := range code.ids {
		binary.Write(buf, binary.LittleEndian, v)
	}

	binary.Write(buf, binary.LittleEndian, int32(len([]byte(code.is_captain))))
	binary.Write(buf, binary.LittleEndian, []byte(code.is_captain))
	return buf.Bytes()
}
func (code Code1) ParseFrom(data []byte) Code1 {
	bufReader := bytes.NewReader(data)

	fmt.Println("--------", bufReader)
	var lenIds int32
	err := binary.Read(bufReader, binary.LittleEndian, &lenIds)
	fmt.Println(err, lenIds)
	var a, b, c int32
	err = binary.Read(bufReader, binary.LittleEndian, &a)
	fmt.Println(err, a)
	err = binary.Read(bufReader, binary.LittleEndian, &b)
	fmt.Println(err, b)
	err = binary.Read(bufReader, binary.LittleEndian, &c)
	fmt.Println(err, c)
	var lenName int32
	err = binary.Read(bufReader, binary.LittleEndian, &lenName)
	fmt.Println(err, lenName)
	var name = make([]byte, lenName, lenName)
	err = binary.Read(bufReader, binary.LittleEndian, &name)
	fmt.Println(err, string(name))
	obj := Code1{}
	obj.SetIds([]int32{a, b, c})
	obj.SetIsCaptain(string(name))

	return obj
}

// func main() {
// 	code1 := Code1{[]int32{1, 2, 3}, "我"}
// 	b := code1.getBytes()
// 	fmt.Printf("%x", b)

// 	code1.parseFrom(b)
// }
