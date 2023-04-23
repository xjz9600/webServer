package recover

import (
	web "routing"
)

type recoverBuilder struct {
	statusCode int
	data       []byte
	log        func(ctx *web.Context)
}

func NewRecoverBuilder(statusCode int, data []byte, log func(ctx *web.Context)) *recoverBuilder {
	return &recoverBuilder{statusCode, data, log}
}

func (r *recoverBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(c *web.Context) {
			defer func() {
				if err := recover(); err != nil {
					c.RespStatusCode = r.statusCode
					c.RespData = r.data
					r.log(c)
				}
			}()
			next(c)
		}
	}
}

func (r *recoverBuilder) Aaa(next web.HandleFunc) web.HandleFunc {
	return func(context *web.Context) {

	}
}
