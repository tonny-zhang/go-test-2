package alarm

import (
	"path/filepath"
	"runtime"
)

// GetError 得到错误
func GetError() {
	pc, filename, line, ok := runtime.Caller(2)

	var functionName string
	if ok {
		functionName = runtime.FuncForPC(pc).Name()
		functionName = filepath.Ext(functionName)
	}

	println(pc, filename, line, ok, functionName)
}
