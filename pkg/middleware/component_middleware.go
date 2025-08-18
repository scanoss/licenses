package middleware

import (
	"context"
	"errors"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/scanoss/papi/api/commonv2"
	"scanoss.com/licenses/pkg/dto"
	"strings"
)

type ComponentMiddleware[T any] struct {
	req *commonv2.ComponentRequest
	MiddlewareBase
}

func NewComponentRequestMiddleware(req *commonv2.ComponentRequest, ctx context.Context) Middleware[dto.ComponentRequestDTO] {
	return &ComponentMiddleware[dto.ComponentRequestDTO]{
		MiddlewareBase: MiddlewareBase{s: ctxzap.Extract(ctx).Sugar()},
		req:            req,
	}
}

func (m *ComponentMiddleware[TOutput]) Process() (dto.ComponentRequestDTO, error) {
	if len(m.req.Purl) == 0 {
		m.s.Warn("no purl request data supplied to decorate. Ignoring request.")
		return dto.ComponentRequestDTO{}, errors.New("no purl request data supplied to decorate")
	}

	purlParts := strings.Split(m.req.Purl, "@")

	if len(purlParts) == 2 {
		m.s.Debugf("PURL split: %s\n", purlParts)
		m.req.Purl = purlParts[0]
		m.req.Requirement = purlParts[1]
	}

	var componentDTO dto.ComponentRequestDTO
	componentDTO.Purl = m.req.Purl
	componentDTO.Requirement = m.req.Requirement

	return componentDTO, nil
}
