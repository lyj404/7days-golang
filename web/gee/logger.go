package gee

import (
	"log"
	"time"
)

func Logger() HandlerFunc {
	return func(context *Context) {
		// 开始时间
		t := time.Now()
		// 处理请求
		context.Next()
		// 计算解决时间
		log.Printf("[%d] %s in %v", context.StatusCode, context.Req.RequestURI, time.Since(t))
	}
}
