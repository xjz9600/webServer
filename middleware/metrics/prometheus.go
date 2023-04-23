package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	web "routing"
	"strconv"
	"time"
)

type prometheusBuilder struct {
	name      string
	subSystem string
	nameSpace string
	help      string
}

func NewPrometheusBuilder(name, subSystem, nameSpace, help string) *prometheusBuilder {
	return &prometheusBuilder{name, subSystem, nameSpace, help}
}

func (p *prometheusBuilder) Build() web.Middleware {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:      p.name,
		Subsystem: p.subSystem,
		Namespace: p.nameSpace,
		Help:      p.help,
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.90:  0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, []string{"pattern", "method", "status"})
	prometheus.MustRegister(vector)
	return func(next web.HandleFunc) web.HandleFunc {
		return func(context *web.Context) {
			startTime := time.Now()
			defer func() {
				duration := time.Now().Sub(startTime).Milliseconds()
				pattern := context.MatchedRoute
				if pattern == "" {
					pattern = "unknown"
				}
				vector.WithLabelValues(pattern, context.Req.Method, strconv.Itoa(context.RespStatusCode)).Observe(float64(duration))
			}()
			next(context)
		}
	}
}
