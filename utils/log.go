package utils

import (
	"io"
	"log"
	"os"
)

var LogFile *os.File

func OpenLogLocal() {
	var err error
	LogFile, err = os.OpenFile("log.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("打开日志文件时出错: %v", err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, LogFile))
}
func CloseLogLocal() {
	if LogFile != nil {
		LogFile.Close()
	}
}
