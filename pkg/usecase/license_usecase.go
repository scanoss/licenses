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
	"strings"
)

type LicenseUseCase struct {
	config                *myconfig.ServerConfig
	licModel              models.LicenseModelInterface
	osadlModel            models.OSADLModelInterface
	componentLicenseModel models.ComponentLicenseModelInterface
	db                    *sqlx.DB
}

func NewLicenseUseCase(config *myconfig.ServerConfig, db *sqlx.DB) *LicenseUseCase {
	return &LicenseUseCase{
		config:                config,
		licModel:              models.NewLicenseModel(db),
		osadlModel:            models.NewOSADLModel(db),
		componentLicenseModel: models.NewComponentLicenseModel(db),
		db:                    db,
	}
}

type Option func(*LicenseUseCase)

// WithLicenseModel option for dependency injection (mainly for testing)
func NewLicenseUseCaseWithLicenseModel(config *myconfig.ServerConfig, licModel models.LicenseModelInterface,
	osadlModel models.OSADLModelInterface, componentLicenseModel models.ComponentLicenseModelInterface) *LicenseUseCase {
	return &LicenseUseCase{
		config:                config,
		licModel:              licModel,
		osadlModel:            osadlModel,
		componentLicenseModel: componentLicenseModel,
	}
}

// GetLicenses
func (lu LicenseUseCase) GetLicenses(ctx context.Context, s *zap.SugaredLogger, components []dto.ComponentRequestDTO) (pb.BasicResponse, *Error) {

	componentLicenses, err := lu.componentLicenseModel.GetComponentLicenses(ctx, s, ""): // TODO: generate MD5 for every purl+version
	if err != nil {
		s.Warn(err)
	}
	s.Debugf("Component Licenses: %v", componentLicenses)

	if len(componentLicenses) == 0 {
		return pb.BasicResponse{}, &Error{Status: common.StatusCode_SUCCEEDED_WITH_WARNINGS, Code: rest.HTTP_CODE_404, Message: "No licenses found", Error: errors.New("no licenses found")}
	}

	var response pb.BasicResponse
	response.Statement = componentLicenses[0].Statement

	var licenses []*pb.BasicLicenseResponse
	for _, cl := range componentLicenses {
		trimmedSPDXIdentifiers := strings.TrimSpace(strings.ReplaceAll(cl.SPDXIdentifiers, " ", ""))
		s.Debugf("SPDX Identifiers: %s", trimmedSPDXIdentifiers)
		// Then split by "/"
		spdxIDS := strings.Split(trimmedSPDXIdentifiers, "/")
		for _, spdxID := range spdxIDS {
			s.Debugf("SPDX ID: %s", spdxID)
			license, errLic := lu.licModel.GetLicenseByID(ctx, s, spdxID)
			if errLic != nil {
				s.Warn(errLic)
			}
			s.Debugf("LICENSE: %v", license)
			licenses = append(licenses, &pb.BasicLicenseResponse{
				Id:       spdxID,
				FullName: license.Name,
			})
		}
	}
	s.Debugf("LICENSES: %v", licenses)
	response.Licenses = licenses
	s.Debugf("RESONSE: %v\n", &response)
	return response, nil
}

// GetDetails
func (lu LicenseUseCase) GetDetails(ctx context.Context, s *zap.SugaredLogger, lic dto.LicenseRequestDTO) (pb.LicenseResponse, *Error) {
	license, err := lu.licModel.GetLicenseByID(ctx, s, lic.ID)
	if err != nil {
		return pb.LicenseResponse{}, &Error{Status: common.StatusCode_FAILED, Code: rest.HTTP_CODE_500, Message: err.Error(), Error: err}
	}
	if license.ID == 0 {
		s.Warnf("License not found: %s", lic.ID)
		return pb.LicenseResponse{}, &Error{Status: common.StatusCode_SUCCEEDED_WITH_WARNINGS, Code: rest.HTTP_CODE_404, Message: "License not found", Error: errors.New("license not found")}
	}
	s.Debugf("License: %v", license)

	osadl, err := lu.osadlModel.GetOSADLByLicenseId(ctx, s, license.LicenseId)
	if err != nil {
		s.Errorf("Error getting OSADL for license: %s, err: %v\n", lic.ID, err)
	}

	s.Debugf("OSADL: %v", osadl)

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
		Osadl: &pb.OSADL{
			Compatibility:          osadl.Compatibilities,
			Incompatibility:        osadl.Incompatibilities,
			CopyleftClause:         osadl.CopyleftClause,
			DependingCompatibility: osadl.DependingCompatibilities,
			PatentHints:            osadl.PatentHints,
		},
	}, nil
}
