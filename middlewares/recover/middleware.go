package recover

import "github.com/nothingZero/mirror"

type MiddlewareBuilder struct {
	StatusCode int
	Data       []byte
	Log        func(ctx *mirror.Context)
}

func (m MiddlewareBuilder) Build() mirror.Middleware {
	return func(next mirror.HandleFunc) mirror.HandleFunc {
		return func(ctx *mirror.Context) {
			defer func() {
				if r := recover(); r != nil {
					ctx.RespData = m.Data
					ctx.RespStatusCode = m.StatusCode
					m.Log(ctx)
				}
			}()
			next(ctx)
		}
	}
}
