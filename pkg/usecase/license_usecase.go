package usecase

import (
	"context"
	"errors"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-models/pkg/scanoss"
	"github.com/scanoss/go-models/pkg/types"
	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/licensesv2"
	"go.uber.org/zap"
	myconfig "scanoss.com/licenses/pkg/config"
	"scanoss.com/licenses/pkg/dto"
	"scanoss.com/licenses/pkg/license"
	models "scanoss.com/licenses/pkg/model"
	"scanoss.com/licenses/pkg/protocol/rest"
)

type LicenseUseCase struct {
	config             *myconfig.ServerConfig
	purlLicenseModel   *models.PurlLicensesModel
	licenseDetailModel models.LicenseDetailModelInterface
	osadlModel         models.OSADLModelInterface
	db                 *sqlx.DB
}

func NewLicenseUseCase(config *myconfig.ServerConfig, db *sqlx.DB) *LicenseUseCase {
	//TODO: refactor scanoss.NewClient to receive only a *sqlx.DB and extract the logger form ctx in all queries.
	// Check: https://scanoss.atlassian.net/browse/SP-3015
	return &LicenseUseCase{
		config:             config,
		licenseDetailModel: models.NewLicenseDetailModel(db),
		purlLicenseModel:   models.NewPurlLicensesModel(db),
		osadlModel:         models.NewOSADLModel(db),
		db:                 db,
	}
}

// WithLicenseModel option for dependency injection (mainly for testing)
func NewLicenseUseCaseWithLicenseModel(config *myconfig.ServerConfig, licenseModel models.LicenseDetailModelInterface,
	osadlModel models.OSADLModelInterface) *LicenseUseCase {
	return &LicenseUseCase{
		config:             config,
		licenseDetailModel: licenseModel,
		osadlModel:         osadlModel,
	}
}

// GetLicenses
func (lu LicenseUseCase) GetLicenses(ctx context.Context, sc *scanoss.Client, crs []dto.ComponentRequestDTO) ([]*pb.ComponentLicenseInfo, *Error) {
	s := ctxzap.Extract(ctx).Sugar()

	var clir []*pb.ComponentLicenseInfo

	// Get a component version. It may be specified on the purl, on the requirement, or the request may not have specified a version at all.
	for _, cr := range crs {

		// Prepare the response so the caller can track each component individually
		componentInfo := &pb.ComponentLicenseInfo{
			Purl:        cr.Purl,
			Requirement: cr.Requirement,
		}
		clir = append(clir, componentInfo)

		c, err := sc.Component.GetComponent(ctx, types.ComponentRequest{
			Purl:        cr.Purl,
			Requirement: cr.Requirement,
		})

		if err != nil {
			s.Warnf("error when resolving component version. %w", err)
			continue
		}

		pl, err := lu.purlLicenseModel.GetLicensesByPurlVersion(ctx, c.Purl, c.Version)
		if err != nil {
			s.Warnf("error when querying purlLicense model for purl=%s version=%s. %w", c.Purl, c.Version, err)
			continue
		}

		if len(pl) == 0 {
			s.Warnf("no license found for purl=%s version=%s. %w", c.Purl, c.Version, err)
			continue
		}

		// Apply source-based priority selection (SPDX tags > attribution files > scancode)
		// TODO: Apply a holistic approach based on all the results.
		bl := license.SelectBestLicense(pl)

		licenseRecord, err := sc.Models.Licenses.GetLicenseByID(ctx, bl.LicenseID)
		if err != nil {
			s.Warnf("error getting license by ID: %d. %v", bl.LicenseID, err)
			continue
		}

		// Parse license expression using SPDX expression parser
		spdxIDs, err := license.ParseLicenseExpression(licenseRecord.SPDX)
		if err != nil {
			s.Warnf("error parsing license expression: %s. %v", licenseRecord.SPDX, err)
			continue
		}

		//TODO: When no license for purl + version
		// * Get the closest version that has license, if the previous and next version match, then we can return that license
		// * Implement fallback to unlicense purl

		// Convert spdxIDs array to []*pb.LicenseInfo
		var licenses []*pb.LicenseInfo
		for _, spdxID := range spdxIDs {
			licenses = append(licenses, &pb.LicenseInfo{
				Id:       spdxID,
				FullName: "",
			})
		}

		componentInfo.Version = c.Version
		componentInfo.Statement = licenseRecord.LicenseName
		componentInfo.Licenses = licenses

	}

	return clir, nil
}

// GetDetails
func (lu LicenseUseCase) GetDetails(ctx context.Context, s *zap.SugaredLogger, lic dto.LicenseRequestDTO) (pb.LicenseDetails, *Error) {
	licenseRecord, err := lu.licenseDetailModel.GetLicenseByID(ctx, s, lic.ID)
	if err != nil {
		return pb.LicenseDetails{}, &Error{Status: common.StatusCode_FAILED, Code: rest.HTTP_CODE_500, Message: err.Error(), Error: err}
	}
	if licenseRecord.ID == 0 {
		s.Warnf("LicenseDetail not found: %s", lic.ID)
		return pb.LicenseDetails{}, &Error{Status: common.StatusCode_SUCCEEDED_WITH_WARNINGS, Code: rest.HTTP_CODE_404, Message: "LicenseDetail not found", Error: errors.New("license not found")}
	}
	s.Debugf("LicenseDetail: %v", licenseRecord)

	osadl, err := lu.osadlModel.GetOSADLByLicenseId(ctx, s, licenseRecord.LicenseId)
	if err != nil {
		s.Errorf("Error getting OSADL for license: %s, err: %v\n", lic.ID, err)
	}

	s.Debugf("OSADL: %v", osadl)

	return pb.LicenseDetails{
		FullName: licenseRecord.Name,
		Spdx: &pb.SPDX{
			FullName:      licenseRecord.Name,
			Id:            licenseRecord.LicenseId,
			DetailsUrl:    licenseRecord.DetailsUrl,
			ReferenceUrl:  licenseRecord.Reference,
			IsDeprecated:  licenseRecord.IsDeprecatedLicenseId,
			IsOsiApproved: licenseRecord.IsOsiApproved,
			SeeAlso:       licenseRecord.SeeAlso,
			IsFsfLibre:    licenseRecord.IsFsfLibre,
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
