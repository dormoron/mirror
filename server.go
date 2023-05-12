package mirror

import (
	"fmt"
	"net"
	"net/http"
)

type HandleFunc func(ctx *Context)

var _ Server = &HTTPServer{}

type Server interface {
	http.Handler
	Start(addr string) error
	registerRoute(method string, path string, handleFunc HandleFunc, ms ...Middleware)
}

type HTTPServerOption func(server *HTTPServer)

type HTTPServer struct {
	router
	mils []Middleware

	log func(msg string, args ...any)
}

func InitHttpServer(opts ...HTTPServerOption) *HTTPServer {
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

// 处理请求入口
func (h *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := &Context{
		Request:        request,
		ResponseWriter: writer,
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

func (h *HTTPServer) Start(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return http.Serve(l, h)
}

func (h *HTTPServer) Get(path string, handleFunc HandleFunc, mils ...Middleware) {
	h.registerRoute(http.MethodGet, path, handleFunc, mils...)
}

func (h *HTTPServer) Head(path string, handleFunc HandleFunc, mils ...Middleware) {
	h.registerRoute(http.MethodHead, path, handleFunc, mils...)
}

func (h *HTTPServer) Post(path string, handleFunc HandleFunc, mils ...Middleware) {
	h.registerRoute(http.MethodPost, path, handleFunc, mils...)
}

func (h *HTTPServer) Put(path string, handleFunc HandleFunc, mils ...Middleware) {
	h.registerRoute(http.MethodPut, path, handleFunc, mils...)
}

func (h *HTTPServer) Patch(path string, handleFunc HandleFunc, mils ...Middleware) {
	h.registerRoute(http.MethodPatch, path, handleFunc, mils...)
}

func (h *HTTPServer) Delete(path string, handleFunc HandleFunc, mils ...Middleware) {
	h.registerRoute(http.MethodDelete, path, handleFunc, mils...)
}

func (h *HTTPServer) Connect(path string, handleFunc HandleFunc, mils ...Middleware) {
	h.registerRoute(http.MethodConnect, path, handleFunc, mils...)
}

func (h *HTTPServer) Options(path string, handleFunc HandleFunc, mils ...Middleware) {
	h.registerRoute(http.MethodOptions, path, handleFunc, mils...)
}

func (h *HTTPServer) Trace(path string, handleFunc HandleFunc, mils ...Middleware) {
	h.registerRoute(http.MethodTrace, path, handleFunc, mils...)
}
