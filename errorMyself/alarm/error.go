package alarm

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// GetError 得到错误
func GetError() {
	pc, filename, line, ok := runtime.Caller(2)

	var functionName string
	if ok {
		functionName = runtime.FuncForPC(pc).Name()
		functionName = filepath.Ext(functionName)
		functionName = strings.TrimPrefix(functionName, ".")
	}

	// println(pc, filename, line, ok, functionName)
	fmt.Printf("%s %s:%d\n", time.Now().Format("2006-01-02 15:04:05"), filename, line)
}
