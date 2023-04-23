package errpage

import (
	web "routing"
)

type errPageBuilder struct {
	resp map[int][]byte
}

func NewErrPageBuilder() *errPageBuilder {
	return &errPageBuilder{
		resp: make(map[int][]byte),
	}
}

func (e *errPageBuilder) AddErrPage(status int, data []byte) *errPageBuilder {
	e.resp[status] = data
	return e
}

func (e *errPageBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(context *web.Context) {
			next(context)
			resp, ok := e.resp[context.RespStatusCode]
			if ok {
				context.RespData = resp
			}
		}
	}
}
