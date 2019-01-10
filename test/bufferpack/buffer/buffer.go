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

			// fmt.Println(nameVal, valData[nameVal])
			write(buf, valData[nameVal], typeVal, itemProp["prop"])
		}
		// elemData := reflect.ValueOf(data).Elem()
		// for _, val := range prop.([]interface{}) {
		// 	itemProp := val.(map[string]interface{})
		// 	typeVal := byte(itemProp["type"].(float64))
		// 	nameVal := itemProp["name"].(string)

		// 	fmt.Println(nameVal, elemData.FieldByName(nameVal))
		// 	fmt.Println(nameVal, elemData.FieldByName(nameVal).CanInterface())
		// 	write(buf, elemData.FieldByName(nameVal).Interface(), typeVal, itemProp["prop"])
		// }
		break
	case typeArray:
		itemProp := prop.(map[string]interface{})
		typeVal := byte(itemProp["type"].(float64))
		fmt.Println(typeVal)
		var lenOfData int
		switch dataData := data.(type) {
		case []byte:
			lenOfData = len(dataData)
			break
		case []int:
			lenOfData = len(dataData)
			break
		case []int8:
			lenOfData = len(dataData)
			break
		case []int16:
			lenOfData = len(dataData)
			break
		case []int32:
			lenOfData = len(dataData)
			break
		case []int64:
			lenOfData = len(dataData)
			break
		case []uint:
			lenOfData = len(dataData)
			break
		case []uint16:
			lenOfData = len(dataData)
			break
		case []uint32:
			lenOfData = len(dataData)
			break
		case []uint64:
			lenOfData = len(dataData)
			break
		case []float32:
			lenOfData = len(dataData)
			break
		case []float64:
			lenOfData = len(dataData)
			break
		case []string:
			lenOfData = len(dataData)
			break
		case []map[string]interface{}:
			lenOfData = len(dataData)

			break
		}
		writeInt(buf, int32(lenOfData))
		fmt.Println(lenOfData)
		for i := 0; i < lenOfData; i++ {
			write(buf, dataData[i], typeVal, nil)
		}
		fmt.Println(lenOfData)
		// valData := data.([]interface{})
		// itemProp := prop.(map[string]interface{})
		// typeVal := byte(itemProp["type"].(float64))

		// lenOfVal := len(data)

		break
	}

	return err
}

// Create 创建buffer
func Create(conf interface{}, data interface{}) ([]byte, error) {
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		// log.Fatal(err)
	// 		fmt.Println("ERROR: ", err)
	// 	}
	// }()
	confObj := conf.(map[string]interface{})

	typeVal := confObj["type"]
	propVal := confObj["prop"]

	buf := new(bytes.Buffer)

	err := write(buf, data, byte(typeVal.(float64)), propVal)

	// if err != nil {
	// 	panic(err)
	// }
	return buf.Bytes(), err
}
