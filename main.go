package main

import (
	"net/http"
	"./gee"
)

func main(){
	router:=gee.Default()
	router.GET("/",func(c *gee.Context) {
		c.String(http.StatusOK,"Hello Gee\n")
	})
	router.GET("/panic",func(c *gee.Context) {
		names:=[]string{"gee"}
		c.String(http.StatusOK,names[100])
	})
	router.Run(":8000")
}