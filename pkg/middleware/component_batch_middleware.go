package middleware

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/scanoss/go-component-helper/componenthelper"
	"github.com/scanoss/papi/api/commonv2"
)

type ComponentBatchMiddleware[TOutput any] struct {
	req *commonv2.ComponentsRequest
	MiddlewareBase
}

func NewComponentsRequestMiddleware(req *commonv2.ComponentsRequest, ctx context.Context) Middleware[[]componenthelper.ComponentDTO] {
	return &ComponentBatchMiddleware[[]componenthelper.ComponentDTO]{
		MiddlewareBase: MiddlewareBase{s: ctxzap.Extract(ctx).Sugar()},
		req:            req,
	}
}

func (m *ComponentBatchMiddleware[TOutput]) Process() ([]componenthelper.ComponentDTO, error) {
	if len(m.req.GetComponents()) == 0 {
		m.s.Warn("No components request data supplied to decorate. Ignoring request.")
		return []componenthelper.ComponentDTO{}, errors.New("no components request data supplied")
	}

	data, err := json.Marshal(m.req.GetComponents())
	if err != nil {
		m.s.Errorf("Problem marshalling dependency request input: %v", err)
		return []componenthelper.ComponentDTO{}, errors.New("problem marshalling request input data")
	}

	var componentDTO []componenthelper.ComponentDTO
	err = json.Unmarshal(data, &componentDTO)
	if err != nil {
		m.s.Errorf("Parse failure: %v", err)
		return []componenthelper.ComponentDTO{}, errors.New("failed to parse request input data")
	}
	return componentDTO, nil
}
