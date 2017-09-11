package mapreduce

import (
	"fmt"
	"log"
	"os"
)

//我的日志
var MyLog *log.Logger

func init() {
	file, err := os.OpenFile("myLog.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error log file : ", err)
	}

	MyLog = log.New(file,
		"MyLog: ", log.Ldate|log.Ltime|log.Lshortfile)

	fmt.Println("日志文件初始化成功！")
}
