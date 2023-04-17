package web

import (
	"net"
	"net/http"
)

// 确保一定实现了server接口
var _ Server = &HTTPServer{}

type HandleFunc func(*Context)

type Server interface {
	http.Handler
	Start(addr string) error
	AddRoute(method string, path string, handleFunc HandleFunc)
	FindRoute(method string, path string) (*matchInfo, bool)
}

type HTTPServer struct {
	router
}

func NewHttpServer() *HTTPServer {
	return &HTTPServer{
		NewRouter()}

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
	h.Serve(ctx)
}

func (h *HTTPServer) Serve(context *Context) {
	n, isFound := h.FindRoute(context.Req.Method, context.Req.URL.Path)
	if !isFound || n.n.handler == nil {
		context.Resp.WriteHeader(404)
		context.Resp.Write([]byte("NOT FOUND"))
		return
	}
	context.PathParams = n.params
	n.n.handler(context)
}

func (h *HTTPServer) Start(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return http.Serve(l, h)
}
