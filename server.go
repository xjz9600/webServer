package web

import (
	"fmt"
	"net"
	"net/http"
)

// 确保一定实现了server接口
var _ Server = &HTTPServer{}

type HandleFunc func(*Context)

type Server interface {
	http.Handler
	Start(addr string) error
	AddRoute(method string, path string, handleFunc HandleFunc, mds ...Middleware)
	FindRoute(method string, path string) (*matchInfo, bool)
}

type HTTPServer struct {
	router
	log func(msg string, args ...any)
	ms  []Middleware
}

type HTTPServerOption func(server *HTTPServer)

func NewHttpServer(opts ...HTTPServerOption) *HTTPServer {
	res := &HTTPServer{
		router: NewRouter(),
		log: func(msg string, args ...any) {
			fmt.Printf(msg, args...)
		},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func ServerWithMiddleware(ms ...Middleware) HTTPServerOption {
	return func(server *HTTPServer) {
		server.ms = ms
	}
}

func (h *HTTPServer) Get(path string, handleFunc HandleFunc) {
	h.AddRoute(http.MethodGet, path, handleFunc)
}

// ServeHTTP 处理请求的入口
func (h *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := &Context{
		Req:  request,
		Resp: writer,
	}
	root := h.Serve
	if len(h.ms) > 0 {
		root = Chain(h.ms...)(root)
	}
	respon := func(next HandleFunc) HandleFunc {
		return func(context *Context) {
			next(context)
			context.Resp.WriteHeader(context.RespStatusCode)
			context.Resp.Write(context.RespData)
		}
	}
	root = respon(root)
	root(ctx)
}

func (h *HTTPServer) Serve(context *Context) {
	n, isFound := h.FindRoute(context.Req.Method, context.Req.URL.Path)
	if !isFound || n.n.handler == nil {
		context.RespStatusCode = http.StatusNotFound
		context.RespData = []byte("NOT FOUND")
		return
	}
	context.PathParams = n.params
	context.MatchedRoute = n.n.route
	Chain(n.mds...)(n.n.handler)(context)
}

func (h *HTTPServer) Start(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return http.Serve(l, h)
}
