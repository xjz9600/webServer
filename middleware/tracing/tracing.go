package tracing

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	web "routing"
)

const instrumentationName = "github.com/xjz9600/webServer/middleware/tracing"

type traceBuilder struct {
	tracer trace.Tracer
}

func (t *traceBuilder) Build() web.Middleware {
	if t.tracer == nil {
		t.tracer = otel.GetTracerProvider().Tracer(instrumentationName)
	}
	return func(next web.HandleFunc) web.HandleFunc {
		return func(context *web.Context) {
			// 尝试跟客户端的 trace 结合在一起
			reqCtx := context.Req.Context()
			reqCtx = otel.GetTextMapPropagator().Extract(reqCtx, propagation.HeaderCarrier(context.Req.Header))
			ctx, span := t.tracer.Start(reqCtx, "unknown")
			defer span.End()

			span.SetAttributes(attribute.String("http.method", context.Req.Method))
			span.SetAttributes(attribute.String("http.utl", context.Req.URL.String()))
			span.SetAttributes(attribute.String("http.host", context.Req.Host))
			context.Req = context.Req.WithContext(ctx)
			next(context)
			span.SetName(context.MatchedRoute)
			span.SetAttributes(attribute.Int("http.status", context.RespStatusCode))
			span.SetAttributes(attribute.String("http.data", string(context.RespData)))
		}
	}
}
