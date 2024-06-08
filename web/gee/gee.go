package gee

import (
	"log"
	"net/http"
	"strings"
)

// HandlerFunc 使用gee的请求处理函数
type HandlerFunc func(*Context)

// RouteGroup 存储路由分组
type (
	RouteGroup struct {
		prefix      string        // 定义了该路由组的基础路径前缀
		middlewares []HandlerFunc // 中间件
		parent      *RouteGroup   // 支持路由组的嵌套，这个字段指向当前路由组的父级路由组
		engin       *Engine       // 指向所有路由组共享的 Engine 实例
	}

	// Engine 定义一个用于存储路由的并实现ServeHTTP的结构体
	Engine struct {
		*RouteGroup // Engine 嵌入了 RouterGroup，从而继承了其所有字段和方法
		router      *router
		groups      []*RouteGroup // 存储所有路由
	}
)

// New 是Engine的构造函数
func New() *Engine {
	engine := &Engine{
		router: newRouter(),
	}
	engine.RouteGroup = &RouteGroup{ // 创建一个新的 RouteGroup 实例，并赋值给 Engine 的 RouteGroup 字段。
		engin: engine, // 将新创建的 Engine 实例赋值给 RouteGroup 的 engine 字段。
	}
	engine.groups = []*RouteGroup{engine.RouteGroup} // 初始化 Engine 的 groups 字段，包含一个 RouteGroup 实例。
	return engine
}

// Group 创建一个新的路由分组，并为其设置前缀。
func (group *RouteGroup) Group(prefix string) *RouteGroup {
	engine := group.engin    // 获取当前路由分组关联的 Engine 实例
	newGroup := &RouteGroup{ // 创建一个新的 RouteGroup 实例
		prefix: group.prefix + prefix, // 新分组的前缀是当前分组的前缀加上新前缀
		parent: group,                 // 新分组的父级设置为当前分组
		engin:  engine,                // 将关联的 Engine 实例赋值给新分组
	}
	engine.groups = append(engine.groups, newGroup) // 将新分组添加到 Engine 的路由分组列表中
	return newGroup
}

// addRoute 添加路由
func (group *RouteGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engin.router.addRoute(method, pattern, handler)
}

// GET 添加GET请求的方法
func (group *RouteGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST 添加POST请求的方法
func (group *RouteGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

// Run 启动HTTP服务器的方法
func (e *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, e)
}

// Use 函数用于向路由分组添加一个或多个中间件。
func (group *RouteGroup) Use(middlewares ...HandlerFunc) {
	// 将传入的中间件追加到分组的中间件切片中
	group.middlewares = append(group.middlewares, middlewares...)
}

// 实现ServeHTTP方法，让所有的请求都交给该实例处理
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 创建一个用于存储中间件处理函数的切片
	var middlewares []HandlerFunc
	// 遍历 Engine 中存储的所有路由分组
	for _, group := range e.groups {
		// 如果请求的 URL 路径以当前分组的前缀开始，说明请求命中了该分组定义的路由
		if strings.HasPrefix(r.URL.Path, group.prefix) {
			// 如果命中，将该分组的中间件追加到中间件切片中
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	// 创建一个新的 Context 实例，包含请求和响应的原始对象
	c := newContext(w, r)
	// 将收集到的中间件设置到 Context 的 handlers 字段
	c.handlers = middlewares
	// 调用 Engine 中的 router 来处理请求，包括执行中间件和查找并执行匹配的路由处理器
	e.router.handle(c)
}
