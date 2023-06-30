package mirror

import (
	"fmt"
	"net"
	"net/http"
)

type HandleFunc func(ctx *Context)

// 确保 HTTPServer 肯定实现了 Server 接口
var _ Server = &HTTPServer{}

type Server interface {
	http.Handler
	// Start 启动服务器
	// addr 是监听地址
	Start(addr string) error
	// registerRoute 注册一个路由
	// method 是 HTTP 方法
	// path 是路由路径
	// handleFunc 是方法
	// mils 是中间件
	registerRoute(method string, path string, handleFunc HandleFunc, mils ...Middleware)
}

type HTTPServerOption func(server *HTTPServer)

type HTTPServer struct {
	router

	mils []Middleware

	log func(msg string, args ...any)

	templateEngine TemplateEngine
}

func InitHTTPServer(opts ...HTTPServerOption) *HTTPServer {
	res := &HTTPServer{
		router: initRouter(),
		log: func(msg string, args ...any) {
			fmt.Printf(msg, args...)
		},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}
func ServerWithTemplateEngine(templateEngine TemplateEngine) HTTPServerOption {
	return func(server *HTTPServer) {
		server.templateEngine = templateEngine
	}
}

func ServerWithMiddleware(mils ...Middleware) HTTPServerOption {
	return func(server *HTTPServer) {
		server.mils = mils
	}
}

func (s *HTTPServer) Use(mils ...Middleware) {
	if s.mils == nil {
		s.mils = mils
		return
	}
	s.mils = append(s.mils, mils...)
}

func (s *HTTPServer) UseRoute(method string, path string, mils ...Middleware) {
	s.registerRoute(method, path, nil, mils...)
}

// ServeHTTP HTTPServer 处理请求的入口
func (h *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := &Context{
		Request:        request,
		ResponseWriter: writer,
		templateEngine: h.templateEngine,
	}
	root := h.server
	for i := len(h.mils) - 1; i >= 0; i-- {
		root = h.mils[i](root)
	}

	var m Middleware = func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			next(ctx)
			h.flashResp(ctx)
		}
	}

	root = m(root)

	root(ctx)
}

func (h *HTTPServer) flashResp(ctx *Context) {
	if ctx.RespStatusCode != 0 {
		ctx.ResponseWriter.WriteHeader(ctx.RespStatusCode)
	}
	write, err := ctx.ResponseWriter.Write(ctx.RespData)
	if err != nil || write != len(ctx.RespData) {
		h.log("写入响应数据失败 %v", err)
	}
}

func (h *HTTPServer) server(ctx *Context) {
	info, ok := h.findRoute(ctx.Request.Method, ctx.Request.URL.Path)
	if !ok || info.n == nil || info.n.handler == nil {
		ctx.RespStatusCode = http.StatusNotFound
		ctx.RespData = []byte("NOT FOUND")
		return
	}
	ctx.PathParams = info.pathParams
	ctx.MatchedRoute = info.n.route
	info.n.handler(ctx)
}

// Start 启动服务器，编程接口
func (h *HTTPServer) Start(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return http.Serve(l, h)
}

func (h *HTTPServer) Get(path string, handleFunc HandleFunc) {
	h.registerRoute(http.MethodGet, path, handleFunc)
}

func (h *HTTPServer) Head(path string, handleFunc HandleFunc) {
	h.registerRoute(http.MethodHead, path, handleFunc)
}

func (h *HTTPServer) Post(path string, handleFunc HandleFunc) {
	h.registerRoute(http.MethodPost, path, handleFunc)
}

func (h *HTTPServer) Put(path string, handleFunc HandleFunc) {
	h.registerRoute(http.MethodPut, path, handleFunc)
}

func (h *HTTPServer) Patch(path string, handleFunc HandleFunc) {
	h.registerRoute(http.MethodPatch, path, handleFunc)
}

func (h *HTTPServer) Delete(path string, handleFunc HandleFunc) {
	h.registerRoute(http.MethodDelete, path, handleFunc)
}

func (h *HTTPServer) Connect(path string, handleFunc HandleFunc) {
	h.registerRoute(http.MethodConnect, path, handleFunc)
}

func (h *HTTPServer) Options(path string, handleFunc HandleFunc) {
	h.registerRoute(http.MethodOptions, path, handleFunc)
}

func (h *HTTPServer) Trace(path string, handleFunc HandleFunc) {
	h.registerRoute(http.MethodTrace, path, handleFunc)
}
