package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Json map[string]interface{}	// 类型别名

// 上下文结构体
type Context struct {
	// 基本http组件
	Writer		http.ResponseWriter
	Req			*http.Request
	// 请求信息
	Path		string
	Method		string
	Params		map[string]string
	// 响应信息
	StatusCode	int
	// 中间件相关
	handlers	[]HandlerFunc	// 存储中间件
	index 		int				// 当前执行都哪个中间件
	engine 		*Engine			// 为了通过engine访问HTML模板
}

func newContext(w http.ResponseWriter,req *http.Request) *Context{
	return &Context{
		Writer:	w,
		Req:	req,
		Path:	req.URL.Path,
		Method: req.Method,
		index:	-1,
	}
}
// 当在中间件中调用Next方法时，控制权交给了下一个中间件直到调用到最后一个中间件。
// 然后再从后往前，调用每个中间件在Next方法之后定义的部分。
func (c *Context) Next() {
	c.index++
	numOfHandlers:=len(c.handlers)
	for ;c.index<numOfHandlers;c.index++{
		c.handlers[c.index](c)	// 依次执行
	}
}

func (c *Context) PostForm(key string) string{
	return c.Req.FormValue(key)	// 返回表单中key对应的value
}
func (c *Context) Query(key string) string{
	return c.Req.URL.Query().Get(key)	// 查询URL附带的参数
}
func (c *Context) Status(code int){
	c.StatusCode=code
	// WriteHeader一般用在往头部写入状态码 必须用在Header.Set之后
	c.Writer.WriteHeader(code)
}
func (c *Context) SetHeader(key string,value string){
	c.Writer.Header().Set(key,value)
}
func (c *Context) Param(key string) string {	// 存储解析得到的参数
	value,_:=c.Params[key]
	return value
}
func (c *Context) Fail(code int,err string) {
	c.index=len(c.handlers)		// 不再处理请求
	c.JSON(code,Json{"message":err})
}

// 构造四种响应
func (c *Context) String(code int,format string,values ...interface{}){
	c.SetHeader("Content-Type","text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format,values...)))
}
func (c *Context) JSON(code int,obj interface{}){
	c.SetHeader("Content-Type","application/json")
	c.Status(code)
	encoder:=json.NewEncoder(c.Writer)	// 创建以Writer为载体的json的encode
	// encoder.Encode: 将obj以json的格式写入到Writer中
	if err:=encoder.Encode(obj);err!=nil{
		http.Error(c.Writer,err.Error(),500)
	}
}
func (c *Context) Data(code int,data []byte){
	c.Status(code)
	c.Writer.Write(data)
}
func (c *Context) HTML(code int,name string,data interface{}){
	// engine.htmlTemplates已经指定了获取模板的目录 所以name只是模板文件名
	c.SetHeader("Content-Type","text/html")
	c.Status(code)
	if err:=c.engine.htmlTemplates.ExecuteTemplate(c.Writer,name,data);err!=nil {
		c.Fail(500,err.Error())
	}
}