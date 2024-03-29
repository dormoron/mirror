package opentelemetry

import (
	"github.com/nothingZero/mirror"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/nothingZero/mirror/middleware/opentelemetry"

type MiddlewareBuilder struct {
	Tracer trace.Tracer
}

func (m MiddlewareBuilder) Build() mirror.Middleware {
	if m.Tracer == nil {
		otel.GetTracerProvider().Tracer(instrumentationName)
	}
	return func(next mirror.HandleFunc) mirror.HandleFunc {
		return func(ctx *mirror.Context) {

			reqCtx := ctx.Request.Context()

			reqCtx = otel.GetTextMapPropagator().Extract(reqCtx, propagation.HeaderCarrier(ctx.Request.Header))

			_, span := m.Tracer.Start(reqCtx, "unknown")

			defer func() {
				span.SetName(ctx.MatchedRoute)

				span.SetAttributes(attribute.Int("http.status", ctx.RespStatusCode))
				span.End()
			}()

			span.SetAttributes(attribute.String("http.method", ctx.Request.Method))
			span.SetAttributes(attribute.String("http.url", ctx.Request.URL.String()))
			span.SetAttributes(attribute.String("http.scheme", ctx.Request.URL.Scheme))
			span.SetAttributes(attribute.String("http.host", ctx.Request.Host))

			ctx.Request = ctx.Request.WithContext(reqCtx)
			next(ctx)
		}
	}
}
