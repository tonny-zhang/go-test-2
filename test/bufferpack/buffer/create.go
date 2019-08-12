package buffer

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const typeInt byte = 0
const typeInt16 byte = 1
const typeFloat byte = 2
const typeDouble byte = 3
const typeString byte = 4
const typeLong byte = 5
const typeByte byte = 6
const typeBool byte = 7
const typeObject byte = 100
const typeArray byte = 101
const typeArrayObject byte = 102
const typeBuffer byte = 103

var keyPrivate = []byte("golang 中注意下面两种字节切片类型的初始化，一个是括号，一个是花括号，其中括号会对字符串内容进行类型转换")
var lenKeyPrivate = len(keyPrivate)

func xor(data []byte) []byte {
	lenData := len(data)
	result := make([]byte, lenData, lenData)

	for i := 0; i < lenData; i++ {
		result[i] = data[i] ^ keyPrivate[i%lenKeyPrivate]
	}
	return result
}
func writeByte(buf *bytes.Buffer, val byte) {
	binary.Write(buf, binary.LittleEndian, val)
}
func writeInt16(buf *bytes.Buffer, val int16) {
	binary.Write(buf, binary.LittleEndian, val)
}
func writeInt(buf *bytes.Buffer, val int32) {
	binary.Write(buf, binary.LittleEndian, val)
}
func writeLong(buf *bytes.Buffer, val int64) {
	binary.Write(buf, binary.LittleEndian, val)
}
func writeFloat(buf *bytes.Buffer, val float32) {
	binary.Write(buf, binary.LittleEndian, val)
}
func writeDouble(buf *bytes.Buffer, val float64) {
	binary.Write(buf, binary.LittleEndian, val)
}
func writeString(buf *bytes.Buffer, val string) {
	lenOfVal := int32(bytes.Count([]byte(val), nil) - 1)
	writeInt(buf, lenOfVal)
	buf.WriteString(val)
}
func write(buf *bytes.Buffer, data interface{}, valType byte, prop interface{}) error {
	var err error
	switch valType {
	case typeByte:
		writeByte(buf, data.(byte))
		break
	case typeInt16:
		writeInt16(buf, data.(int16))
		break
	case typeInt:
		var int32Val int32
		switch data.(type) {
		case byte:
			int32Val = int32(data.(byte))
			break
		case int16:
			int32Val = int32(data.(int16))
			break
		case int32:
			int32Val = int32(data.(int32))
			break
		default:
			err = fmt.Errorf("[%s] not int val", data)
		}
		if err == nil {
			writeInt(buf, int32Val)
		}
		break
	case typeLong:
		var int64Val int64
		switch data.(type) {
		case byte:
			int64Val = int64(data.(byte))
			break
		case int16:
			int64Val = int64(data.(int16))
			break
		case int32:
			int64Val = int64(data.(int32))
			break
		case int64:
			int64Val = int64(data.(int64))
			break
		default:
			err = fmt.Errorf("[%s] not long val", data)
		}
		if err == nil {
			writeLong(buf, int64Val)
		}
		break
	case typeFloat:
		writeFloat(buf, data.(float32))
		break
	case typeDouble:
		var doubleVal float64
		switch data.(type) {
		case byte:
			doubleVal = float64(data.(byte))
			break
		case int16:
			doubleVal = float64(data.(int16))
			break
		case int32:
			doubleVal = float64(data.(int32))
			break
		case int64:
			doubleVal = float64(data.(int64))
			break
		case float32:
			doubleVal = float64(data.(float32))
			break
		case float64:
			doubleVal = float64(data.(float64))
			break
		default:
			err = fmt.Errorf("[%s] not double val", data)
		}
		if err == nil {
			writeDouble(buf, doubleVal)
		}

		break
	case typeString:
		var strVal string
		switch data.(type) {
		case string:
			strVal = data.(string)
			break
		default:
			err = fmt.Errorf("[%s] not string val", data)
		}
		if err == nil {
			writeString(buf, strVal)
		}
		break
	case typeBool:
		var boolVal bool
		switch data.(type) {
		case bool:
			boolVal = data.(bool)
			break
		default:
			if data != nil && data != 0 {
				boolVal = true
			}
		}

		if boolVal {
			writeByte(buf, 1)
		} else {
			writeByte(buf, 0)
		}

		break
	case typeObject:
		valData := data.(map[string]interface{})
		for _, val := range prop.([]interface{}) {
			itemProp := val.(map[string]interface{})
			typeVal := byte(itemProp["type"].(float64))
			nameVal := itemProp["name"].(string)

			write(buf, valData[nameVal], typeVal, itemProp["prop"])
		}
		break
	case typeArray:
		itemProp := prop.(map[string]interface{})
		typeVal := byte(itemProp["type"].(float64))
		switch typeVal {
		case typeByte:
			d := data.([]byte)
			writeInt(buf, int32(len(d)))
			for i := 0; i < len(d); i++ {
				writeByte(buf, d[i])
			}
		case typeInt16:
			d := data.([]int16)
			writeInt(buf, int32(len(d)))
			for i := 0; i < len(d); i++ {
				writeInt16(buf, d[i])
			}
		case typeInt:
			d := data.([]int32)
			writeInt(buf, int32(len(d)))
			for i := 0; i < len(d); i++ {
				writeInt(buf, d[i])
			}
		case typeLong:
			d := data.([]int64)
			writeInt(buf, int32(len(d)))
			for i := 0; i < len(d); i++ {
				writeLong(buf, d[i])
			}
		case typeFloat:
			d := data.([]float32)
			writeInt(buf, int32(len(d)))
			for i := 0; i < len(d); i++ {
				writeFloat(buf, d[i])
			}
		case typeDouble:
			d := data.([]float64)
			writeInt(buf, int32(len(d)))
			for i := 0; i < len(d); i++ {
				writeDouble(buf, d[i])
			}
		}
		break
	case typeArrayObject:
		itemProp := prop.([]interface{})
		d := data.([]map[string]interface{})
		writeInt(buf, int32(len(itemProp)))
		for i := 0; i < len(itemProp); i++ {
			write(buf, d[i], typeObject, itemProp)
		}
	}

	return err
}

// Create 创建buffer, 并得到私钥加密后的数据
func Create(conf interface{}, data interface{}) ([]byte, error) {
	defer func() {
		if err := recover(); err != nil {
			// log.Fatal(err)
			fmt.Println("ERROR: ", err)
		}
	}()
	confObj := conf.(map[string]interface{})

	typeVal := confObj["type"]
	propVal := confObj["prop"]

	buf := new(bytes.Buffer)

	err := write(buf, data, byte(typeVal.(float64)), propVal)

	// if err != nil {
	// 	panic(err)
	// }
	// if err == nil {
	// 	return xor(buf.Bytes()), err
	// } else {
	// 	return buf.Bytes(), err
	// }

	// fmt.Printf("old bytes : %x\n", buf.Bytes())
	// fmt.Printf("xor bytes : %x\n", xor(buf.Bytes()))
	// fmt.Printf("xor2 bytes : %x\n", xor(xor(buf.Bytes())))
	return xor(buf.Bytes()), err
}
