package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/scanoss/papi/api/commonv2"
	"scanoss.com/licenses/pkg/dto"
	"strings"
)

type ComponentBatchMiddleware[TOutput any] struct {
	req *commonv2.ComponentBatchRequest
	MiddlewareBase
}

func NewComponentBatchMiddleware(req *commonv2.ComponentBatchRequest, ctx context.Context) Middleware[[]dto.ComponentRequestDTO] {
	return &ComponentBatchMiddleware[[]dto.ComponentRequestDTO]{
		MiddlewareBase: MiddlewareBase{s: ctxzap.Extract(ctx).Sugar()},
		req:            req,
	}
}

func (m *ComponentBatchMiddleware[TOutput]) groupComponentsByPurl(c []dto.ComponentRequestDTO) map[string]dto.ComponentRequestDTO {
	componentMap := make(map[string]dto.ComponentRequestDTO)
	for _, comp := range c {
		key := comp.Purl
		if comp.Requirement != "" {
			key += "@" + comp.Requirement
		}
		// Handle requests with purl@version
		splitPurl := strings.Split(comp.Purl, "@")
		if len(splitPurl) >= 2 {
			comp = dto.ComponentRequestDTO{
				Purl:        splitPurl[0],
				Requirement: splitPurl[1],
			}
		}
		componentMap[key] = comp
	}
	return componentMap
}

func (m *ComponentBatchMiddleware[TOutput]) Process() ([]dto.ComponentRequestDTO, error) {

	if len(m.req.GetComponents()) == 0 {
		m.s.Warn("No components request data supplied to decorate. Ignoring request.")
		return []dto.ComponentRequestDTO{}, errors.New("no components request data supplied")
	}

	data, err := json.Marshal(m.req.GetComponents())
	if err != nil {
		m.s.Errorf("Problem marshalling dependency request input: %v", err)
		return []dto.ComponentRequestDTO{}, errors.New("problem marshalling request input data")
	}

	var componentEntity []dto.ComponentRequestDTO
	err = json.Unmarshal(data, &componentEntity)
	if err != nil {
		m.s.Errorf("Parse failure: %v", err)
		return nil, errors.New("failed to parse request input data")
	}

	componentMap := m.groupComponentsByPurl(componentEntity)

	// Convert map to slice
	components := make([]dto.ComponentRequestDTO, 0, len(componentMap))
	for _, component := range componentMap {
		components = append(components, component)
	}
	m.s.Debugf("components: %v\n", components)
	m.s.Debugf("%v to process\n", len(components))
	return components, nil
}
