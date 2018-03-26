package raft

import "log"
import (
	"fmt"
	"os"
)

// Debugging
const Debug = 0

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		log.Printf(format, a...)
	}
	return
}

//我的日志
var ZM *log.Logger

func init() {
	file, err := os.OpenFile("ZM.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error log file : ", err)
	}

	ZM = log.New(file,
		"ZM: ", log.Ldate|log.Ltime|log.Lshortfile)

	fmt.Println("日志文件初始化成功！")
}
