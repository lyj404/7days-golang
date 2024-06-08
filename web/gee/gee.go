package gee

import (
	"html/template"
	"log"
	"net/http"
	"path"
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
		*RouteGroup                      // Engine 嵌入了 RouterGroup，从而继承了其所有字段和方法
		router        *router            // 存储路由
		groups        []*RouteGroup      // 存储所有路由
		htmlTemplates *template.Template // 用于存储 HTML 模板的编译结果
		funcMap       template.FuncMap   // 定义了 HTML 模板渲染时可以使用的自定义函数映射
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

// Default 默认使用日志和错误恢复中间件
func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
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

// createStaticHandler 创建一个处理静态文件服务的 HandlerFunc
func (group *RouteGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	// 将分组的前缀与相对路径连接，形成静态文件的绝对路径
	absolutePath := path.Join(group.prefix, relativePath)
	// 使用 http.StripPrefix 包装 http.FileServer，创建一个文件服务器处理器
	// 这个处理器将去掉请求 URL 中的前缀部分，然后将其传递给文件系统
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	// 返回一个 HandlerFunc，它使用创建的文件服务器来处理请求。
	return func(c *Context) {
		// 从 URL 参数中获取文件路径。
		file := c.Param("filepath")
		// 检查文件是否存在以及我们是否有权限访问它。
		if _, err := fs.Open(file); err != nil {
			// 如果文件打开失败（文件不存在或无权限），设置状态码为 404 并返回。
			c.Status(http.StatusNotFound)
			return
		}
		// 使用文件服务器处理器来响应请求。
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// Static 方法用于注册一个静态文件服务的路由
func (group *RouteGroup) Static(relativePath string, root string) {
	// 使用 createStaticHandler 方法创建一个静态文件服务处理器
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	// 定义 URL 模式匹配相对路径下的所有文件请求
	// 使用路径通配符 `*filepath` 来捕获 URL 中的剩余部分作为文件路径
	urlPattern := path.Join(relativePath, "/*filepath")
	group.GET(urlPattern, handler)
}

// SetFuncMap 方法用于设置 HTML 模板渲染时使用的自定义函数映射
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

// LoadHTMLGlob 方法用于加载和解析匹配特定模式的 HTML 模板文件
func (engine *Engine) LoadHTMLGlob(pattern string) {
	// 使用模板包的 Must 方法来创建一个新的模板，使用 Engine 的 funcMap 作为函数映射
	// ParseGlob 方法用于加载匹配指定模式（如 "*.html"）的所有文件，并将它们解析为模板
	// 使用 template.Must 来检查解析过程中是否有错误发生，如果有，将导致程序崩溃并输出错误信息
	// 方法内部使用 template.New("") 创建一个新的模板，然后通过 .Funcs(engine.funcMap) 将自定义函数映射应用到模板上
	// .ParseGlob(pattern) 调用 template.ParseGlob 方法，根据提供的模式加载 HTML 文件，并解析为模板
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
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
	// 将当前的 Engine 实例赋值给 Context 的 engine 字段，这样 Context 就可以访问 Engine 提供的功能
	c.engine = e
	// 调用 Engine 中的 router 来处理请求，包括执行中间件和查找并执行匹配的路由处理器
	e.router.handle(c)
}
