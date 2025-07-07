package middleware

import (
	pb "github.com/scanoss/papi/api/licensesv2"
	"go.uber.org/zap"
	"scanoss.com/licenses/pkg/interfaces"
)

type LicenseDetailMiddleware[TOutput any] struct {
	req *pb.LicenseRequest
	MiddlewareBase
}

func NewLicenseDetailMiddleware(req *pb.LicenseRequest, s *zap.SugaredLogger) Middleware[[]interfaces.Component] {
	return &LicenseDetailMiddleware[[]interfaces.Component]{
		req:            req,
		MiddlewareBase: MiddlewareBase{s: s},
	}
}

func (m *LicenseDetailMiddleware[TOutput]) Process() ([]interfaces.Component, error) {
	return nil, nil
}
