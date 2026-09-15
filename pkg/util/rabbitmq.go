package util

import (
	"log"
	"strings"
)

// go 获取命令行参数，第零个元素一般是程序名，输入的参数从下标1开始
func BodyFrom(args []string) string {
	var s string
	if (len(args) < 2) || args[1] == "" {
		s = "hello"
	} else {
		s = strings.Join(args[1:], " ")
	}
	return s
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
