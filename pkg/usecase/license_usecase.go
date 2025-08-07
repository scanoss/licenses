package usecase

import (
	"context"
	"errors"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	purlutils "github.com/scanoss/go-purl-helper/pkg"
	"strings"

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
	config       *myconfig.ServerConfig
	licenseModel models.LicenseDetailModelInterface
	osadlModel   models.OSADLModelInterface
	db           *sqlx.DB
}

func NewLicenseUseCase(config *myconfig.ServerConfig, db *sqlx.DB) *LicenseUseCase {
	return &LicenseUseCase{
		config:       config,
		licenseModel: models.NewLicenseDetailModel(db),
		osadlModel:   models.NewOSADLModel(db),
		db:           db,
	}
}

type Option func(*LicenseUseCase)

// WithLicenseModel option for dependency injection (mainly for testing)
func NewLicenseUseCaseWithLicenseModel(config *myconfig.ServerConfig, licenseModel models.LicenseDetailModelInterface,
	osadlModel models.OSADLModelInterface) *LicenseUseCase {
	return &LicenseUseCase{
		config:       config,
		licenseModel: licenseModel,
		osadlModel:   osadlModel,
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

		p, err := purlutils.PurlFromString(c.Purl)
		if err != nil {
			s.Warnf("error parsing purl: %s. %w", c.Purl, err)
		}

		urls, err := sc.Models.AllUrls.GetURLsByPurlNameTypeVersion(ctx, p.Name, p.Type, c.Version)
		var licenseID int32

		if err != nil || len(urls) == 0 {
			s.Warnf("AllUrls query failed for purl=%s version=%s - trying fallback (error: %v)", c.Purl, c.Version, err)

			// Fallback: try to get license from LDB component licenses table
			s.Debugf("Trying fallback with LDB component licenses for purl: %s", c.Purl)

			lclm := models.NewLDBComponentLicensesModel(lu.db)
			purlMD5 := lclm.CalculateMD5FromPurlVersion(c.Purl, c.Version)
			s.Debugf("Calculated PURL MD5: %s for purl=%s version=%s", purlMD5, c.Purl, c.Version)
			componentLicenses, ldbErr := lclm.GetLicensesByPurlMD5(ctx, purlMD5)

			if ldbErr != nil {
				s.Warnf("fallback failed - cannot get license from LDB component licenses: %v", ldbErr)
				continue
			}

			if len(componentLicenses) == 0 {
				s.Warnf("no license found in fallback for purl=%s - version=%s", c.Purl, c.Version)
				continue
			}

			licenseID = componentLicenses[0].LicenseID
			s.Debugf("Found license via fallback: licenseID=%d for purl=%s", licenseID, c.Purl)
		} else {
			licenseID = urls[0].LicenseID
		}

		license, err := sc.Models.Licenses.GetLicenseByID(ctx, licenseID)
		if err != nil {
			s.Warnf("error getting license by ID: %d. %v", licenseID, err)
			continue
		}

		spdxIDs := strings.Split(license.LicenseID, "/")

		// Convert spdxIDs array to []*pb.LicenseInfo
		var licenses []*pb.LicenseInfo
		for _, spdxID := range spdxIDs {
			licenses = append(licenses, &pb.LicenseInfo{
				Id:       strings.TrimSpace(spdxID),
				FullName: "", // TODO: Implement SPDX ID to full name mapping
			})
		}

		componentInfo.Version = c.Version
		componentInfo.Statement = license.LicenseName
		componentInfo.Licenses = licenses

	}

	return clir, nil
}

// GetDetails
func (lu LicenseUseCase) GetDetails(ctx context.Context, s *zap.SugaredLogger, lic dto.LicenseRequestDTO) (pb.LicenseDetails, *Error) {
	license, err := lu.licenseModel.GetLicenseByID(ctx, s, lic.ID)
	if err != nil {
		return pb.LicenseDetails{}, &Error{Status: common.StatusCode_FAILED, Code: rest.HTTP_CODE_500, Message: err.Error(), Error: err}
	}
	if license.ID == 0 {
		s.Warnf("LicenseDetail not found: %s", lic.ID)
		return pb.LicenseDetails{}, &Error{Status: common.StatusCode_SUCCEEDED_WITH_WARNINGS, Code: rest.HTTP_CODE_404, Message: "LicenseDetail not found", Error: errors.New("license not found")}
	}
	s.Debugf("LicenseDetail: %v", license)

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
