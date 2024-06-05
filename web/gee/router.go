package gee

import (
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

// parsePattern 函数用于解析路由模式，返回路由模式的各个部分组成的字符串数组
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/") // 将路由模式按斜杠分割成字符串数组

	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item) // 将非空部分添加到 parts 数组中
			if item[0] == '*' {
				break // 如果遇到通配符，停止遍历
			}
		}
	}
	return parts
}

// addRoute 方法用于向路由器中添加路由规则及其对应的处理函数
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePattern(pattern) // 解析路由模式，得到路由模式的各个部分组成的字符串数组
	key := method + "-" + pattern  // 构造路由规则的唯一标识

	_, ok := r.roots[method] // 检查当前 HTTP 方法是否已经存在根节点
	if !ok {
		r.roots[method] = &node{} // 如果不存在，则创建一个新的根节点
	}
	r.roots[method].insert(pattern, parts, 0) // 在根节点上插入新的路由规则
	r.handlers[key] = handler                 // 将路由规则与对应的处理函数关联起来
}

// getRoute 方法用于根据请求的方法和路径，从路由器中查找匹配的路由规则
func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path) // 解析请求路径，得到路径的各个部分组成的字符串数组
	params := make(map[string]string) // 存储路由中的参数

	root, ok := r.roots[method] // 获取请求方法对应的根节点

	if !ok {
		return nil, nil // 如果根节点不存在，则返回空
	}

	n := root.search(searchParts, 0) // 在根节点上搜索匹配的路由规则节点

	if n != nil {
		parts := parsePattern(n.pattern) // 解析匹配的路由规则，得到规则的各个部分组成的字符串数组
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index] // 将通配符参数与实际路径参数进行映射
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/") // 处理通配符参数后的额外路径参数
				break
			}
		}
		return n, params
	}

	return nil, nil // 返回匹配的路由规则节点以及参数映射
}

func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern
		r.handlers[key](c)
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}
