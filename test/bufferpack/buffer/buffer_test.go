package buffer

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"testing"
)

func call(t *testing.T, confJSON string, data interface{}, base64Expect string, method string) {
	var conf interface{}
	err := json.Unmarshal([]byte(confJSON), &conf)
	if err != nil {
		log.Fatal(err)
	}
	result, err := Create(conf, data)
	resultBase64 := base64.StdEncoding.EncodeToString(result)
	if resultBase64 != base64Expect {
		t.Errorf("%s result is %s; want %s", method, resultBase64, base64Expect)
	}
}
func TestCreateInt16(t *testing.T) {
	var confJSON = `{"type":1, "prop": [{"type": 2, "name": "name"}]}`
	var int16Val int16 = 1
	call(t, confJSON, int16Val, "Zm8=", "createInt")
}

func TestCreateByte(t *testing.T) {
	var confJSON = `{"type":6}`
	var val byte = 1
	call(t, confJSON, val, "Zg==", "createByte")
}

func TestCreateInt(t *testing.T) {
	var confJSON = `{"type": 0}`
	var intVal int32 = 1
	call(t, confJSON, intVal, "Zm9sYQ==", "createInt")
}
func TestCreateIntUseInt16(t *testing.T) {
	var confJSON = `{"type": 0}`
	var intVal int16 = 1
	call(t, confJSON, intVal, "Zm9sYQ==", "createIntUseInt16")
}
func TestCreateLong(t *testing.T) {
	var confJSON = `{"type": 5}`
	var intVal int64 = 100
	call(t, confJSON, intVal, "A29sYW5nIOQ=", "createIntUseInt16")
}
func TestCreateFloat(t *testing.T) {
	var confJSON = `{"type": 2}`
	var val float32 = 100
	call(t, confJSON, val, "Z2+kIw==", "createFloat")
}
func TestCreateDouble(t *testing.T) {
	var confJSON = `{"type": 3}`
	var val float64 = 100
	call(t, confJSON, val, "Z29sYW5neaQ=", "createDouble")
}
