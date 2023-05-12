package errhdl

import "mirror"

type MiddlewareBuilder struct {
	resp map[int][]byte
}

func InitMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{resp: make(map[int][]byte)}
}
func (m *MiddlewareBuilder) AddCode(status int, data []byte) *MiddlewareBuilder {
	m.resp[status] = data
	return m
}

func (m MiddlewareBuilder) Build() mirror.Middleware {
	return func(next mirror.HandleFunc) mirror.HandleFunc {
		return func(ctx *mirror.Context) {
			next(ctx)
			resp, ok := m.resp[ctx.RespStatusCode]
			if ok {
				ctx.RespData = resp
			}
		}
	}
}
