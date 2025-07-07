package middleware

import (
	"encoding/json"
	"errors"
	"github.com/scanoss/papi/api/commonv2"
	"go.uber.org/zap"
	"scanoss.com/licenses/pkg/interfaces"
	"strings"
)

type ComponentBatchMiddleware[TOutput any] struct {
	req *commonv2.ComponentBatchRequest
	MiddlewareBase
}

func NewComponentBatchMiddleware(req *commonv2.ComponentBatchRequest, s *zap.SugaredLogger) Middleware[[]interfaces.Component] {
	return &ComponentBatchMiddleware[[]interfaces.Component]{
		MiddlewareBase: MiddlewareBase{s: s},
		req:            req,
	}
}

func (m *ComponentBatchMiddleware[TOutput]) groupComponentsByPurl(c []interfaces.Component) map[string]interfaces.Component {
	componentMap := make(map[string]interfaces.Component)
	for _, comp := range c {
		key := comp.Purl
		if comp.Requirement != "" {
			key += "@" + comp.Requirement
		}
		// Handle requests with purl@version
		splitPurl := strings.Split(comp.Purl, "@")
		if len(splitPurl) >= 2 {
			comp = interfaces.Component{
				Purl:        splitPurl[0],
				Requirement: splitPurl[1],
			}
		}
		componentMap[key] = comp
	}
	return componentMap
}

func (m *ComponentBatchMiddleware[TOutput]) Process() ([]interfaces.Component, error) {

	if len(m.req.GetComponents()) == 0 {
		m.s.Warn("No components request data supplied to decorate. Ignoring request.")
		return nil, errors.New("no components request data supplied")
	}

	data, err := json.Marshal(m.req.GetComponents())
	if err != nil {
		m.s.Errorf("Problem marshalling dependency request input: %v", err)
		return nil, errors.New("problem marshalling request input data")
	}

	var componentEntity []interfaces.Component
	err = json.Unmarshal(data, &componentEntity)
	if err != nil {
		m.s.Errorf("Parse failure: %v", err)
		return nil, errors.New("failed to parse request input data")
	}

	componentMap := m.groupComponentsByPurl(componentEntity)

	// Convert map to slice
	components := make([]interfaces.Component, 0, len(componentMap))
	for _, component := range componentMap {
		components = append(components, component)
	}
	m.s.Debugf("components: %v\n", components)
	m.s.Debugf("%v to process\n", len(components))
	return components, nil
}
