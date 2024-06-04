package gee

import (
	"net/http"
)

// HandlerFunc 使用gee的请求处理函数
type HandlerFunc func(*Context)

// Engine 定义一个用于存储路由的并实现ServeHTTP的结构体
type Engine struct {
	router *router
}

// New 是Engine的构造函数
func New() *Engine {
	return &Engine{
		router: newRouter(),
	}
}

// addRoute 添加路由
func (e *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	e.router.addRoute(method, pattern, handler)
}

// GET 添加GET请求的方法
func (e *Engine) GET(pattern string, handler HandlerFunc) {
	e.addRoute("GET", pattern, handler)
}

// POST 添加POST请求的方法
func (e *Engine) POST(pattern string, handler HandlerFunc) {
	e.addRoute("POST", pattern, handler)
}

// Run 启动HTTP服务器的方法
func (e *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, e)
}

// 实现ServeHTTP方法，让所有的请求都交给该实例处理
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := newContext(w, r)
	e.router.handle(c)
}
