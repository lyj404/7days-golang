package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

// Context 封装了 HTTP 请求和响应的上下文信息
type Context struct {
	// 原始对象
	Writer http.ResponseWriter
	Req    *http.Request
	// 请求信息
	Path   string
	Method string
	Params map[string]string
	// 响应信息
	StatusCode int
	// 中间件
	handlers []HandlerFunc // 中间件处理函数的切片
	index    int           // 当前处理的中间件索引
}

// newContext 创建一个新的 Context 实例
func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    r,
		Path:   r.URL.Path,
		Method: r.Method,
		index:  -1, // 初始化中间件索引为 -1，表示尚未开始处理中间件
	}
}

func (c *Context) Next() {
	c.index++            // 将中间件索引向前移动到下一个中间件
	s := len(c.handlers) // 获取中间件切片的长度
	for ; c.index < s; c.index++ {
		// 调用当前索引处的中间件处理函数
		// 当 c.index < s 时循环调用中间件，直到所有中间件都执行完毕
		c.handlers[c.index](c)
	}
}

func (c *Context) Fail(code int, err string) {
	// 将 c.index 设置为 len(c.handlers)，这表示已经执行完所有中间件和处理器
	// 这样做可以确保 Next() 方法在被调用时不会尝试执行任何后续的处理函数
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

// PostForm 返回指定 key 的表单值
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

// Query 返回指定 key 的查询参数值
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

// Status 设置响应的状态码
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

// SetHeader 设置响应头信息
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

// String 以纯文本格式发送响应
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

// JSON 以 JSON 格式发送响应
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

// Data 发送原始字节数据作为响应
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

// HTML 以 HTML 格式发送响应
func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}
