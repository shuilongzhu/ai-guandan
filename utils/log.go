package utils

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
)

func CallerInfo(skip int) string {
	rpc := make([]uintptr, 1)
	//skip 1:本方法信息；2:本方法的调用者方法信息；3:本方法的调用者方法的调用者方法信息
	n := runtime.Callers(skip, rpc[:])
	if n < 1 {
		return "-"
	}
	frame, _ := runtime.CallersFrames(rpc).Next()
	filePath := strings.ReplaceAll(frame.File, projectRootPath(), "")
	funcName := strings.Split(frame.Function, ".")[1]
	return fmt.Sprintf("%s:%d method:%s()", filePath, frame.Line, funcName)
}

func projectRootPath() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir+"/", "\\", "/", -1)
}
