package web

import (
	"net"
	"net/http"
)

// 确保一定实现了server接口
var _ Server = &HTTPServer{}

type HandleFunc func(Context)

type Server interface {
	http.Handler
	Start(addr string) error
	AddRoute(method string, path string, handleFunc HandleFunc)
}

type HTTPServer struct {
}

func (h *HTTPServer) AddRoute(method string, path string, handleFunc HandleFunc) {
	//panic("implement me")
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

}

func (h *HTTPServer) Start(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return http.Serve(l, h)
}
