package middleware

import (
	"encoding/json"
	"errors"
	pb "github.com/scanoss/papi/api/licensesv2"
	"go.uber.org/zap"
	"scanoss.com/licenses/pkg/dto"
)

type LicenseDetailMiddleware[TOutput any] struct {
	req *pb.LicenseRequest
	MiddlewareBase
}

func NewLicenseDetailMiddleware(req *pb.LicenseRequest, s *zap.SugaredLogger) Middleware[dto.LicenseRequestDTO] {
	return &LicenseDetailMiddleware[dto.LicenseRequestDTO]{
		req:            req,
		MiddlewareBase: MiddlewareBase{s: s},
	}
}

func (m *LicenseDetailMiddleware[TOutput]) Process() (dto.LicenseRequestDTO, error) {

	if len(m.req.GetId()) == 0 {
		m.s.Warn("No license request data supplied to decorate. Ignoring request.")
		return dto.LicenseRequestDTO{}, errors.New("no license request data supplied")
	}

	data, err := json.Marshal(m.req)
	if err != nil {
		m.s.Errorf("Problem marshalling dependency request input: %v", err)
		return dto.LicenseRequestDTO{}, errors.New("problem marshalling request input data")
	}

	var licenseDTO dto.LicenseRequestDTO
	err = json.Unmarshal(data, &licenseDTO)
	if err != nil {
		m.s.Errorf("Parse failure: %v", err)
		return dto.LicenseRequestDTO{}, errors.New("failed to parse request input data")
	}

	return licenseDTO, nil
}
