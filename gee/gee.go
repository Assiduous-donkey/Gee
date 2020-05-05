package gee

import (
	"log"
	"strings"
	"net/http"
	"html/template"
	"path"
)

// 路由处理函数
type HandlerFunc func(*Context)

// 分组控制
type RouterGroup struct {
	prefix		string
	middlewares	[]HandlerFunc	// 支持的中间件
	parent		*RouterGroup
	engine		*Engine			// 使分组具有添加路由的功能
}

// 实现ServerHttp方法的路由引擎
type Engine struct{
	// 整个框架的资源都由Engine统一协调 所以Engine中嵌套了RouterGroup
	*RouterGroup
	router 			*router
	groups 			[]*RouterGroup		// 存储所有的分组
	// 用于模板渲染
	htmlTemplates	*template.Template	// 用于将模板加载到内存
	funcMap			template.FuncMap	// 自定义模板渲染函数
}

func New() *Engine{
	engine:=&Engine{router:newRouter()}
	engine.RouterGroup=&RouterGroup{engine:engine}
	engine.groups=[]*RouterGroup{engine.RouterGroup}
	return engine
}
func Default() *Engine {
	// 默认使用logger和recovery这两个中间件
	engine:=New()
	engine.Use(middlewares.Logger(),middlewares.Recovery())
	return engine
}

// 创建(子)分组
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine:=group.engine		// 整个服务自始至终都只有这一个engine
	newGroup:=&RouterGroup{
		prefix:group.prefix+prefix,
		parent:group,
		engine:engine,
	}
	engine.groups=append(engine.groups,newGroup)
	return newGroup
}

// 注册路由
func (group *RouterGroup) addRoute(method,comp string,handler HandlerFunc){
	pattern:=group.prefix+comp
	log.Printf("Route %4s - %s",method,pattern)
	group.engine.router.addRoute(method,pattern,handler)
}

// 先实现简单的GET和POST（添加路由）
// 由于Engine继承了RouterGroup 所以也可以直接通过Engine添加路由
func (group *RouterGroup) GET(pattern string,handler HandlerFunc){
	group.addRoute("GET",pattern,handler)
}
func (group *RouterGroup) POST(pattern string,handler HandlerFunc){
	group.addRoute("POST",pattern,handler)
} 

// 将中间件应用到某个Group
func (group *RouterGroup) Use(middlewares ...HandlerFunc){
	// 采用的是可变数量的参数
	group.middlewares=append(group.middlewares,middlewares...)
}

// 运行服务器
func (engine *Engine) ServeHTTP(w http.ResponseWriter,req *http.Request){
	var middlewares []HandlerFunc
	// 接收到一个具体的请求时要判断该请求适用于那些中间件(可以简单的由URL前缀来判断)
	for _,group:=range engine.groups {
		if strings.HasPrefix(req.URL.Path,group.prefix) {
			middlewares=append(middlewares,group.middlewares...)
		}
	}
	// 每个请求对应一个上下文Context
	c:=newContext(w,req)
	c.handlers=middlewares
	c.engine=engine
	engine.router.handle(c)
}
func (engine *Engine) Run(addr string) (err error){
	// 实际执行的是ServeHTTP方法
	return http.ListenAndServe(addr,engine)
}

// 处理静态资源
func (group *RouterGroup) createStaticHandler(relativePath string,fs http.FileSystem) HandlerFunc{
	// FileSystem是一个接口 包含Open方法
	// absolutePath实际上是访问静态文件的URL前缀
	// absolutePath+静态文件所在的路径 才等于 请求静态文件的URL
	absolutePath:=path.Join(group.prefix,relativePath)
	// http.FileServer返回一个handler用来处理访问fs所代表的目录的HTTP请求
	// http.StripPrefix用于过滤掉URL的前缀以获取具体的静态文件名
	fileServer:=http.StripPrefix(absolutePath,http.FileServer(fs))
	return func(c *Context) {
		file:=c.Param("filepath")
		// 检查文件是否存在以及是否有访问文件的权限
		if _,err:=fs.Open(file);err!=nil {
			c.Status(http.StatusNotFound)
			return 
		}
		fileServer.ServeHTTP(c.Writer,c.Req)
	}
}
// 暴露给用户的接口 用户可以将磁盘上某个目录root映射到路由relativePath
func (group *RouterGroup) Static(relativePath,root string) {
	// root指的是静态文件的根目录
	// relativePath是静态文件对应的URL
	// http.Dir(root)的作用是利用本地目录root生成一个文件系统
	handler:=group.createStaticHandler(relativePath,http.Dir(root))
	// 将relativePath和/*filepath拼接成动态路由 并加入到GET方法中
	urlPattern:=path.Join(relativePath,"/*filepath")
	group.GET(urlPattern,handler)
}

// 模板渲染相关
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap=funcMap
}
func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates=template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}