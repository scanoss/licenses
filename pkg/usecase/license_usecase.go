package usecase

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-models/pkg/scanoss"
	"github.com/scanoss/go-models/pkg/types"
	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/licensesv2"
	"go.uber.org/zap"
	myconfig "scanoss.com/licenses/pkg/config"
	"scanoss.com/licenses/pkg/dto"
	models "scanoss.com/licenses/pkg/model"
	"scanoss.com/licenses/pkg/protocol/rest"
)

type LicenseUseCase struct {
	config     *myconfig.ServerConfig
	licModel   models.LicenseModelInterface
	osadlModel models.OSADLModelInterface
	db         *sqlx.DB
}

func NewLicenseUseCase(config *myconfig.ServerConfig, db *sqlx.DB) *LicenseUseCase {
	return &LicenseUseCase{
		config:     config,
		licModel:   models.NewLicenseModel(db),
		osadlModel: models.NewOSADLModel(db),
		db:         db,
	}
}

type Option func(*LicenseUseCase)

// WithLicenseModel option for dependency injection (mainly for testing)
func NewLicenseUseCaseWithLicenseModel(config *myconfig.ServerConfig, licModel models.LicenseModelInterface,
	osadlModel models.OSADLModelInterface) *LicenseUseCase {
	return &LicenseUseCase{
		config:     config,
		licModel:   licModel,
		osadlModel: osadlModel,
	}
}

// GetLicenses
func (lu LicenseUseCase) GetLicenses(ctx context.Context, s *zap.SugaredLogger, sc *scanoss.Client, components []dto.ComponentRequestDTO) ([]*pb.ComponentLicenseInfo, *Error) {

	for _, c := range components {
		_, _ = sc.Component.GetComponent(types.ComponentRequest{
			Purl:        c.Purl,
			Requirement: c.Requirement,
		})

	}

	var a []*pb.ComponentLicenseInfo

	a = append(a, &pb.ComponentLicenseInfo{
		Purl:      "example",
		Version:   "v1.0.0",
		Statement: "GPL",
		Licenses: []*pb.LicenseInfo{
			{Id: "GPL v2", FullName: "General Public Licence V2"},
		},
	})

	return a, nil
}

// GetDetails
func (lu LicenseUseCase) GetDetails(ctx context.Context, s *zap.SugaredLogger, lic dto.LicenseRequestDTO) (pb.LicenseDetails, *Error) {
	license, err := lu.licModel.GetLicenseByID(ctx, s, lic.ID)
	if err != nil {
		return pb.LicenseDetails{}, &Error{Status: common.StatusCode_FAILED, Code: rest.HTTP_CODE_500, Message: err.Error(), Error: err}
	}
	if license.ID == 0 {
		s.Warnf("License not found: %s", lic.ID)
		return pb.LicenseDetails{}, &Error{Status: common.StatusCode_SUCCEEDED_WITH_WARNINGS, Code: rest.HTTP_CODE_404, Message: "License not found", Error: errors.New("license not found")}
	}
	s.Debugf("License: %v", license)

	osadl, err := lu.osadlModel.GetOSADLByLicenseId(ctx, s, license.LicenseId)
	if err != nil {
		s.Errorf("Error getting OSADL for license: %s, err: %v\n", lic.ID, err)
	}

	s.Debugf("OSADL: %v", osadl)

	return pb.LicenseDetails{
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
		Osadl: &pb.OSADL{
			Compatibility:          osadl.Compatibilities,
			Incompatibility:        osadl.Incompatibilities,
			CopyleftClause:         osadl.CopyleftClause,
			DependingCompatibility: osadl.DependingCompatibilities,
			PatentHints:            osadl.PatentHints,
		},
	}, nil
}
