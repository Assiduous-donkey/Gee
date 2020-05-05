package gee

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
	"../gee"
)

func Recovery() gee.HandlerFunc {
	return func(c *gee.Context) {
		// recovery只能作用在defer中
		// panic会导致程序被中止但是在退出前会先处理完当前协程上已经defer的任务然后再退出
		// 当panic被触发时控制权被交给了defer
		defer func() {
			if err:=recover();err!=nil {
				// 将err转为字符串类型
				message:=fmt.Sprintf("%s",err)
				log.Printf("%s\n\n",trace(message))
				c.Fail(http.StatusInternalServerError,"Internal Server Error")
			}
		}()
		c.Next()
	}
}
func trace(message string) string {
	var pcs [32]uintptr
	// Callers用来返回调用栈的程序计数器
	// 第0个是Callers 第1个是trace 第2个是Recovery中的defer func ...
	// 所以跳掉前3个
	n:=runtime.Callers(3,pcs[:])
	var str strings.Builder
	str.WriteString(message+"\nTraceback:")
	for _,pc:=range pcs[:n] {
		fn:=runtime.FuncForPC(pc)	// 根据程序计数器获取对应的函数
		file,line:=fn.FileLine(pc)	// 获取调用该函数的文件名和行号
		str.WriteString(fmt.Sprintf("\n\t%s:%d",file,line))
	}
	return str.String()
}