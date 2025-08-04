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
	licenseModel models.LicenseModelInterface
	osadlModel   models.OSADLModelInterface
	db           *sqlx.DB
}

func NewLicenseUseCase(config *myconfig.ServerConfig, db *sqlx.DB) *LicenseUseCase {
	return &LicenseUseCase{
		config:       config,
		licenseModel: models.NewLicenseModel(db),
		osadlModel:   models.NewOSADLModel(db),
		db:           db,
	}
}

type Option func(*LicenseUseCase)

// WithLicenseModel option for dependency injection (mainly for testing)
func NewLicenseUseCaseWithLicenseModel(config *myconfig.ServerConfig, licenseModel models.LicenseModelInterface,
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

	var componentInfos []*pb.ComponentLicenseInfo

	// Get a component version. It may be specified on the purl, on the requirement, or the request may not have specified a version at all.
	for _, cr := range crs {

		// Prepare the response so the caller can track each component individually
		componentInfo := &pb.ComponentLicenseInfo{
			Purl:        cr.Purl,
			Requirement: cr.Requirement,
		}

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
		if err != nil {
			s.Warnf("cannot get url from purl: %s. %w", c.Purl, err)
			continue
		}

		if len(urls) == 0 {
			s.Warnf("no license found for purl=%s - version=%s. %w", c.Purl, c.Version)
			continue
		}

		// Asume that all URLs have same license. One URL is one .tar.gz on a Github Release for example
		// Sometimes a purl@version have multiple urls (because multiple sources released)
		// TODO: In order to solve this mapping issue we need to mine licenses based on purl@version
		//		There is an intent to solve this mapping with the table ldb_component_licenses but there are so many missing licenses
		licenseID := urls[0].LicenseID

		license, err := sc.Models.Licenses.GetLicenseByID(ctx, licenseID)
		if err != nil {
			s.Warnf("error getting license by ID: %d. %v", licenseID, err)
			continue
		}

		// license.LicenseName: contains the license as declared from the project. There are scenarios where a project may declare multiple licenses.
		//					usually is done with SPDX expressions, but some projects declare csv of licenses, or separated by slash

		// license.LicenseID: This is the SPDX version of the license name, if the project have multiple licenses then, each is separated by a "/"

		// There are others scenarios where the project does not declare a specific string license name. and instead have multiple licenses
		// that apply based on the component compiled or component used. Those cases are not considerer here, like FFmpeg/FFmpeg

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
		componentInfos = append(componentInfos, componentInfo)

	}

	return componentInfos, nil
}

// GetDetails
func (lu LicenseUseCase) GetDetails(ctx context.Context, s *zap.SugaredLogger, lic dto.LicenseRequestDTO) (pb.LicenseDetails, *Error) {
	license, err := lu.licenseModel.GetLicenseByID(ctx, s, lic.ID)
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
