package web

type Middleware func(next HandleFunc) HandleFunc

func Chain(m ...Middleware) Middleware {
	return func(next HandleFunc) HandleFunc {
		for i := len(m) - 1; i >= 0; i-- {
			next = m[i](next)
		}
		return next
	}
}
