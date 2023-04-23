package logging

import (
	"encoding/json"
	web "routing"
)

type logBuilder struct {
	logFunc func(log string)
}

func NewLog(fu func(log string)) *logBuilder {
	return &logBuilder{logFunc: fu}
}

func (l *logBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(context *web.Context) {
			defer func() {
				log := accessLog{
					Host:       context.Req.Host,
					Route:      context.MatchedRoute,
					HTTPMethod: context.Req.Method,
					Path:       context.Req.URL.Path,
				}
				data, _ := json.Marshal(log)
				l.logFunc(string(data))
			}()
			next(context)
		}
	}
}

type accessLog struct {
	Host       string `json:"host,omitempty"`
	Route      string `json:"route,omitempty"`
	HTTPMethod string `json:"http_method,omitempty"`
	Path       string `json:"path,omitempty"`
}
