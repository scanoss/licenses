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
	"strings"
)

type LicenseUseCase struct {
	config             *myconfig.ServerConfig
	sc                 *scanoss.Client
	purlLicenseModel   *models.PurlLicensesModel
	licenseDetailModel models.LicenseDetailModelInterface
	osadlModel         models.OSADLModelInterface
	db                 *sqlx.DB
}

func NewLicenseUseCase(config *myconfig.ServerConfig, db *sqlx.DB) *LicenseUseCase {
	return &LicenseUseCase{
		config:             config,
		sc:                 scanoss.New(db),
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

// GetComponentLicense
func (lu LicenseUseCase) GetComponentLicense(ctx context.Context, crs dto.ComponentRequestDTO) (*pb.ComponentLicenseInfo, *Error) {
	// Reuse existing GetComponentsLicense logic with single-item array
	results, err := lu.GetComponentsLicense(ctx, []dto.ComponentRequestDTO{crs})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return &pb.ComponentLicenseInfo{}, nil
	}
	return results[0], nil
}

// GetComponentsLicense
func (lu LicenseUseCase) GetComponentsLicense(ctx context.Context, crs []dto.ComponentRequestDTO) ([]*pb.ComponentLicenseInfo, *Error) {
	s := ctxzap.Extract(ctx).Sugar()

	var clir []*pb.ComponentLicenseInfo

	for _, cr := range crs {

		// Prepare the response so the caller can track each component individually
		// Determine which PURL to use, original format if it was split from purl@version
		purl := cr.Purl
		if cr.OriginalPurl != "" {
			purl = cr.OriginalPurl
		}

		// Determine requirement, empty for split PURLs
		requirement := cr.Requirement
		if cr.WasSplit {
			requirement = ""
		}

		componentInfo := &pb.ComponentLicenseInfo{
			Purl:        purl,
			Requirement: requirement,
		}
		clir = append(clir, componentInfo)

		c, err := lu.sc.Component.GetComponent(ctx, types.ComponentRequest{
			Purl:        cr.Purl,
			Requirement: cr.Requirement,
		})

		if err != nil {
			s.Warnf("error when resolving component version. %w", err)
			continue
		}

		purlLicenses, err := lu.purlLicenseModel.GetLicensesByPurlVersionAndSource(ctx, c.Purl, c.Version, []int16{
			license.SourceComponentDeclared,
			license.SourceSPDXAttributionFiles,
			license.SourceInternalAttributionFiles})

		if err != nil {
			s.Warnf("error when querying GetVersionedAndUnversionedLicenses() for purl=%s version=%s: %v", c.Purl, c.Version, err)
			continue
		}

		if len(purlLicenses) == 0 {
			s.Info("no purlLicenses data found for purl=%s version=%s. Trying with unversioned purl", c.Purl, c.Version)
			purlLicenses, err = lu.purlLicenseModel.GetLicensesByUnversionedPurlAndSource(ctx, c.Purl, []int16{
				license.SourceComponentDeclared,
				license.SourceSPDXAttributionFiles,
				license.SourceInternalAttributionFiles})

			if len(purlLicenses) == 0 {
				s.Info("no purlLicenses data found for unversioned purl=%s.", c.Purl, c.Version)
				continue
			}

			if err != nil {
				s.Warnf("error when querying GetLicensesByUnversionedPurlAndSource() for purl=%s: %v", c.Purl, err)
				continue
			}

		}

		//Retrieve all the uniques licenses ids
		dedupLicensesIDs := license.ExtractLicenseIDsFromPurlLicenses(purlLicenses)
		if len(dedupLicensesIDs) == 0 {
			s.Warnf("no license data available for purl=%s version=%s", c.Purl, c.Version)
			continue
		}

		s.Debugf("Found %d unique license_ids from all sources for purl=%s version=%s", len(dedupLicensesIDs), c.Purl, c.Version)

		var finalLicenses []*pb.LicenseInfo

		// Process ALL license_ids with SPDX-level
		allSpdxLicenses := make(map[string]bool)
		for _, licenseID := range dedupLicensesIDs {
			licenseRecord, err := lu.sc.Models.Licenses.GetLicenseByID(ctx, licenseID)
			if err != nil {
				s.Warnf("error getting license by ID: %d. %v", licenseID, err)
				continue
			}

			// Parse license expression using SPDX expression parser
			spdx, err := license.ParseLicenseExpression(licenseRecord.SPDX)
			if err != nil {
				s.Warnf("error parsing license expression for license_id %d: %s. %v", licenseID, licenseRecord.SPDX, err)
				continue
			}

			// Add each SPDX license to our deduplicated collection
			for _, l := range spdx {

				spdxLicenseDetail, err := lu.sc.Models.Licenses.GetSPDXLicenseDetails(ctx, l)
				if err != nil {
					s.Warnf("error getting SPDX license details for license_id %d: %s. %v", licenseID, licenseRecord.SPDX, err)
					spdxLicenseDetail.Name = ""
				}

				if !allSpdxLicenses[l] {
					allSpdxLicenses[l] = true
					finalLicenses = append(finalLicenses, &pb.LicenseInfo{
						Id:       l,
						FullName: spdxLicenseDetail.Name,
					})
				}
			}
		}

		// If no licenses could be processed, log and continue
		if len(finalLicenses) == 0 {
			s.Warnf("no valid licenses found after processing %d license IDs for purl=%s version=%s", len(dedupLicensesIDs), c.Purl, c.Version)
			continue
		}

		// Build statement by joining all license IDs with " AND "
		var licenseIDs []string
		for _, l := range finalLicenses {
			licenseIDs = append(licenseIDs, l.Id)
		}
		statement := strings.Join(licenseIDs, " AND ")

		componentInfo.Version = c.Version
		componentInfo.Statement = statement
		componentInfo.Licenses = finalLicenses

	}

	return clir, nil
}

// GetDetails
func (lu LicenseUseCase) GetDetails(ctx context.Context, s *zap.SugaredLogger, lic dto.LicenseRequestDTO) (pb.LicenseDetails, *Error) {
	licenseRecord, err := lu.sc.Models.Licenses.GetSPDXLicenseDetails(ctx, lic.ID)
	if err != nil {
		return pb.LicenseDetails{}, &Error{Status: common.StatusCode_FAILED, Code: rest.HTTP_CODE_500, Message: err.Error(), Error: err}
	}
	if licenseRecord.ID == "" {
		s.Warnf("LicenseDetail not found: %s", lic.ID)
		return pb.LicenseDetails{}, &Error{Status: common.StatusCode_SUCCEEDED_WITH_WARNINGS, Code: rest.HTTP_CODE_404, Message: "LicenseDetail not found", Error: errors.New("license not found")}
	}
	s.Debugf("LicenseDetail: %v", licenseRecord)

	//TODO: Add OSADL model to postgress db
	/*	osadl, err := lu.osadlModel.GetOSADLByLicenseId(ctx, s, licenseRecord.LicenseId)
		if err != nil {
			s.Errorf("Error getting OSADL for license: %s, err: %v\n", lic.ID, err)
		}

		s.Debugf("OSADL: %v", osadl)*/

	return pb.LicenseDetails{
		FullName: licenseRecord.Name,
		Spdx: &pb.SPDX{
			FullName:      licenseRecord.Name,
			Id:            licenseRecord.ID,
			DetailsUrl:    licenseRecord.DetailsURL,
			ReferenceUrl:  licenseRecord.Reference,
			IsDeprecated:  licenseRecord.IsDeprecatedLicenseId,
			IsOsiApproved: licenseRecord.IsOsiApproved,
			SeeAlso:       licenseRecord.SeeAlso,
		},
		// TODO: Add OSADL model to postgress db
		/*		Osadl: &pb.OSADL{
				Compatibility:          osadl.Compatibilities,
				Incompatibility:        osadl.Incompatibilities,
				CopyleftClause:         osadl.CopyleftClause,
				DependingCompatibility: osadl.DependingCompatibilities,
				PatentHints:            osadl.PatentHints,
			},*/
	}, nil
}
