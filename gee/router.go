package gee

import (
	"net/http"
	"strings"
)

type router struct {
	roots		map[string]*node		// 为了实现动态路由而需要的记录
	handlers	map[string]HandlerFunc	
}

// 辅助的方法
func parsePattern(pattern string) []string {
	// 以'/'为分割符划分URL
	path:=strings.Split(pattern,"/")
	parts:=make([]string,0)
	for _,item:=range path {
		if item!=""{
			parts=append(parts,item)
			// URL中的“*”必须是URL的最后一个匹配节点
			if item[0]=='*'{
				break
			}
		}
	}
	return parts
}

func newRouter() *router{
	return &router{
		roots:		make(map[string]*node),
		handlers:	make(map[string]HandlerFunc),
	}
}
func (r *router) addRoute(method,pattern string,handler HandlerFunc){
	parts:=parsePattern(pattern)
	path:=method+"-"+pattern
	if _,ok:=r.roots[method];ok==false {
		r.roots[method]=&node{}
	}
	r.roots[method].insert(pattern,parts,0)
	r.handlers[path]=handler
}
// 返回的map存储了解析的结果
func (r *router) getRoute(method,path string) (*node,map[string]string) {
	searchParts:=parsePattern(path)		// 动态路由的parts
	params:=make(map[string]string)		// URL动态匹配 匹配到的参数
	root,ok:=r.roots[method]
	if !ok {
		return nil,nil
	}
	node:=root.search(searchParts,0)	// 路由匹配
	if node!=nil {	// 该URL存在
		parts:=parsePattern(node.pattern) // 实际匹配到的路由的parts
		/*********************************
		匹配的例子:
		/gee/go/demo匹配到/gee/:lang/demo 解析结果为{lang:"go"}
		/gee/go/demo匹配到/grr/*path 解析结果为{path:"go/demo"}
		**********************************/
		for index,part:=range parts {
			if part[0]==':' {	// 获取动态匹配到的参数
				params[part[1:]]=searchParts[index]
			}
			if part[0]=='*'&&len(part)>1 {
				params[part[1:]]=strings.Join(searchParts[index:],"/")
				break
			}
		}
		return node,params
	}
	return nil,nil
}
func (r *router) handle(c *Context){
	node,params:=r.getRoute(c.Method,c.Path)
	if node!=nil{
		c.Params=params
		path:=c.Method+"-"+node.pattern
		// 将路由匹配成功得到的Handler添加到该请求要执行的中间件后面
		c.handlers=append(c.handlers,r.handlers[path])
	} else {
		c.handlers=append(c.handlers,func(c *Context){
			c.String(http.StatusNotFound,"404 not found: %s\n",c.Path)
		})
	}
	c.Next()	// 依次执行 中间件 - 处理函数 - 中间件
}