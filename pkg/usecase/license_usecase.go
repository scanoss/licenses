package usecase

import (
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/licensesv2"
	"go.uber.org/zap"
	myconfig "scanoss.com/licenses/pkg/config"
	"scanoss.com/licenses/pkg/dto"
	models "scanoss.com/licenses/pkg/model"
	"scanoss.com/licenses/pkg/protocol/rest"
)

type LicenseUseCase struct {
	s        *zap.SugaredLogger
	ctx      context.Context
	config   *myconfig.ServerConfig
	licModel models.LicenseModelInterface
	db       *sqlx.DB
}

func NewLicenseUseCase(ctx context.Context, s *zap.SugaredLogger, config *myconfig.ServerConfig, db *sqlx.DB) *LicenseUseCase {
	return &LicenseUseCase{
		s:        s,
		ctx:      ctx,
		config:   config,
		licModel: models.NewLicenseModel(ctx, s, db),
		db:       db,
	}
}

type Option func(*LicenseUseCase)

// WithLicenseModel option for dependency injection (mainly for testing)
func NewLicenseUseCaseWithLicenseModel(ctx context.Context, s *zap.SugaredLogger, config *myconfig.ServerConfig, model models.LicenseModelInterface) *LicenseUseCase {
	return &LicenseUseCase{
		ctx:      ctx,
		s:        s,
		config:   config,
		licModel: model,
	}
}

// GetLicenses
func (lu LicenseUseCase) GetLicenses(ctx context.Context, components []dto.ComponentRequestDTO) {

}

// GetDetails
func (lu LicenseUseCase) GetDetails(lic dto.LicenseRequestDTO) (pb.LicenseResponse, *Error) {
	license, err := lu.licModel.GetLicenseByID(lic.ID)
	if err != nil {
		return pb.LicenseResponse{}, &Error{Status: common.StatusCode_FAILED, Code: rest.HTTP_CODE_500, Message: err.Error(), Error: err}
	}
	if license.ID == 0 {
		lu.s.Warnf("License not found: %s", lic.ID)
		return pb.LicenseResponse{}, &Error{Status: common.StatusCode_SUCCEEDED_WITH_WARNINGS, Code: rest.HTTP_CODE_404, Message: "License not found", Error: errors.New("license not found")}
	}
	lu.s.Debugf("License: %v", license)
	return pb.LicenseResponse{
		FullName: license.Name,
		Spdx: &pb.SPDX{
			FullName:      license.Name,
			Id:            license.LicenseId,
			DetailsUrl:    license.DetailsUrl,
			ReferenceUrl:  license.Reference,
			IsDeprecated:  license.IsDeprecatedLicenseId,
			IsOsiApproved: license.IsOsiApproved,
			SeeAlso:       license.SeeAlso,
			IsFsfLibre:    license.IsFsfLibre,
		},
	}, nil
}
