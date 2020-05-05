package gee

import (
	"fmt"
	"testing"
	"reflect"
)

func newTestRouter() *router {
	r:=newRouter()
	r.addRoute("GET","/hello",nil)
	r.addRoute("GET","/hello/:name",nil)
	r.addRoute("GET","/hello/b/c",nil)
	r.addRoute("GET","/bi/:name",nil)
	r.addRoute("GET","/asserts/*filepath",nil)
	return r
}

func TestParsePattern(t *testing.T) {
	ok:=reflect.DeepEqual(parsePattern("/p/:name"),[]string{"p",":name"})
	ok=ok&&reflect.DeepEqual(parsePattern("/p/*"), []string{"p", "*"})
	ok=ok&&reflect.DeepEqual(parsePattern("/p/*name/*"), []string{"p", "*name"})
	if !ok {
		t.Fatal("test parsePattern failed")
	} else {
		fmt.Println("parsePattern successfully")
	}
}

func TestGetRoute(t *testing.T) {
	r:=newTestRouter()
	node,_:=r.getRoute("GET","/hello")
	if node==nil {
		t.Fatal("there are no urls matched")
	}
	if node.pattern!="/hello" {
		t.Fatal("should match /hello/:name")
	}
	fmt.Printf("matched path: %s\n", node.pattern)
}