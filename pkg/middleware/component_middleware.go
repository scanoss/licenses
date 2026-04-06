package middleware

import (
	"context"
	"errors"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/scanoss/go-component-helper/componenthelper"
	"github.com/scanoss/papi/api/commonv2"
)

type ComponentMiddleware[T any] struct {
	req *commonv2.ComponentRequest
	MiddlewareBase
}

func NewComponentRequestMiddleware(req *commonv2.ComponentRequest, ctx context.Context) Middleware[componenthelper.ComponentDTO] {
	return &ComponentMiddleware[componenthelper.ComponentDTO]{
		MiddlewareBase: MiddlewareBase{s: ctxzap.Extract(ctx).Sugar()},
		req:            req,
	}
}

func (m *ComponentMiddleware[TOutput]) Process() (componenthelper.ComponentDTO, error) {
	if len(m.req.Purl) == 0 {
		m.s.Warn("no purl request data supplied to decorate. Ignoring request.")
		return componenthelper.ComponentDTO{}, errors.New("no purl request data supplied to decorate")
	}

	var componentDTO componenthelper.ComponentDTO
	componentDTO.Purl = m.req.Purl
	componentDTO.Requirement = m.req.Requirement

	return componentDTO, nil
}
