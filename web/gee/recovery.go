package gee

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

func trace(message string) string {
	// pcs 是一个切片，用于存储调用者程序计数器的值
	var pcs [32]uintptr
	// runtime.Callers 填充 pcs 切片，记录调用者的信息
	// 参数 3 表示跳过前 3 层调用者，即不包括 trace 函数和它的直接调用者
	n := runtime.Callers(3, pcs[:])

	// 创建一个字符串构建器
	var str strings.Builder
	// 写入传入的消息和 "Traceback:" 到构建器
	str.WriteString(message + "\nTraceback:")
	// 遍历 pcs 切片中的每个调用者程序计数器的值。
	for _, pc := range pcs[:n] {
		// runtime.FuncForPC 返回给定程序计数器 pc 的函数信息
		fn := runtime.FuncForPC(pc)
		// 返回当前 pc 所在的文件名和行号
		file, line := fn.FileLine(pc)
		// 格式化字符串并写入构建器，包含文件名和行号
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	// 返回构建器中存储的堆栈跟踪字符串
	return str.String()
}

func Recovery() HandlerFunc {
	// 返回一个 HandlerFunc，它内部定义了一个匿名函数，用于创建一个中间件。
	return func(c *Context) {
		// 使用 defer 来延迟执行，确保即使发生 panic，也能执行 panic 处理逻辑
		defer func() {
			// recover 用于捕获当前 goroutine 的 panic 值，并返回 nil 如果没有 panic 发生
			if err := recover(); err != nil {
				// 将 panic 值转换为字符串
				message := fmt.Sprintf("%s", err)
				// 打印堆栈跟踪到日志，包括 panic 消息和堆栈跟踪
				log.Printf("%s\n\n", trace(message))
				// 在 Context 上调用 Fail 方法，返回 HTTP 500 状态码和默认错误消息给客户端
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()

		// 调用 Context 的 Next 方法来继续执行链中的下一个中间件或处理函数
		c.Next()
	}
}
