package gee

import (
	"log"
	"time"
	"../gee"
)

func Logger()  gee.HandlerFunc{
	return func(c *gee.Context) {
		t:=time.Now()
		c.Next()
		// %v表示输出值的默认表示
		log.Printf("[%d] %s in %v",c.StatusCode,c.Req.RequestURI,time.Since(t))
	}
}