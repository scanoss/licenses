package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-component-helper/componenthelper"
	comphelputils "github.com/scanoss/go-component-helper/componenthelper/utils"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	"github.com/scanoss/go-models/pkg/scanoss"
	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/licensesv2"
	"go.uber.org/zap"
	"net/http"
	"scanoss.com/licenses/pkg/cache"
	myconfig "scanoss.com/licenses/pkg/config"
	"scanoss.com/licenses/pkg/dto"
	"scanoss.com/licenses/pkg/license"
	models "scanoss.com/licenses/pkg/model"
	"strings"
)

type LicenseUseCase struct {
	config             *myconfig.ServerConfig
	sc                 *scanoss.Client
	purlLicenseModel   *models.PurlLicensesModel
	licenseDetailModel models.LicenseDetailModelInterface
	osadlModel         models.OSADLModelInterface
	spdxLicenseCache   cache.SPDXLicenseCacheInterface
	db                 *sqlx.DB
}

func NewLicenseUseCase(config *myconfig.ServerConfig, db *sqlx.DB, spdxCache cache.SPDXLicenseCacheInterface) *LicenseUseCase {
	return &LicenseUseCase{
		config:             config,
		sc:                 scanoss.New(db),
		licenseDetailModel: models.NewLicenseDetailModel(db),
		purlLicenseModel:   models.NewPurlLicensesModel(db),
		osadlModel:         models.NewOSADLModel(db),
		spdxLicenseCache:   spdxCache,
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
func (lu LicenseUseCase) GetComponentLicense(ctx context.Context, dto componenthelper.ComponentDTO) (*pb.ComponentLicenseInfo, *Error) {
	// Reuse existing GetComponentsLicense logic with single-item array
	results, err := lu.GetComponentsLicense(ctx, []componenthelper.ComponentDTO{dto})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return &pb.ComponentLicenseInfo{}, nil
	}
	return results[0], nil
}

// GetComponentsLicense
func (lu LicenseUseCase) GetComponentsLicense(ctx context.Context, componentDTOs []componenthelper.ComponentDTO) ([]*pb.ComponentLicenseInfo, *Error) {
	s := ctxzap.Extract(ctx).Sugar()

	var clir []*pb.ComponentLicenseInfo

	processedComponents := componenthelper.GetComponentsVersion(componenthelper.ComponentVersionCfg{
		MaxWorkers: 5,
		DB:         lu.db,
		Ctx:        ctx,
		S:          s,
		Input:      componentDTOs,
	})

	var validComponents []componenthelper.Component
	for _, c := range processedComponents {
		if c.Status.StatusCode != domain.Success && c.Status.StatusCode != domain.VersionNotFound {
			clir = append(clir, &pb.ComponentLicenseInfo{
				Purl:         c.OriginalPurl,
				Requirement:  c.OriginalRequirement,
				Version:      c.Version,
				ComponentUrl: c.URL,
				ErrorMessage: &c.Status.Message,
				ErrorCode:    domain.StatusCodeToErrorCode(c.Status.StatusCode),
			})
			continue
		}
		validComponents = append(validComponents, c)
	}

	for _, c := range validComponents {

		componentInfo := &pb.ComponentLicenseInfo{
			Purl:         c.OriginalPurl,
			Requirement:  c.OriginalRequirement,
			ComponentUrl: c.URL,
		}

		version := c.Version
		if c.Version == "" && c.Requirement != "" && len(c.Versions) > 0 {
			version = comphelputils.FindNearestVersion(c.Requirement, c.Versions)
			msg := fmt.Sprintf("Version not found, using nearest version %s", version)
			componentInfo.ErrorMessage = &msg
			componentInfo.ErrorCode = domain.StatusCodeToErrorCode(domain.VersionNotFound)
		}
		componentInfo.Version = version

		purlLicenses, err := lu.purlLicenseModel.GetLicensesByPurlVersionAndSource(ctx, c.Purl, version, []int16{
			license.SourceComponentDeclared,
			license.SourceSPDXAttributionFiles,
			license.SourceInternalAttributionFiles})

		if err != nil {
			s.Warnf("error when querying GetVersionedAndUnversionedLicenses() for purl=%s version=%s: %v", c.Purl, c.Version, err)
			s.Warnf("error when querying GetVersionedAndUnversionedLicenses() for purl=%s version=%s: %v", c.Purl, version, err)
			message := fmt.Sprintf("License info not found for %s", c.Purl)
			componentInfo.ErrorMessage = &message
			componentInfo.ErrorCode = domain.StatusCodeToErrorCode(domain.ComponentWithoutInfo)
			clir = append(clir, componentInfo)
			continue
		}

		if len(purlLicenses) == 0 {
			s.Info("no purlLicenses data found for purl=%s version=%s. Trying with unversioned purl", c.Purl, version)
			purlLicenses, err = lu.purlLicenseModel.GetLicensesByUnversionedPurlAndSource(ctx, c.Purl, []int16{
				license.SourceComponentDeclared,
				license.SourceSPDXAttributionFiles,
				license.SourceInternalAttributionFiles})

			if err != nil {
				s.Warnf("error when querying GetLicensesByUnversionedPurlAndSource() for purl=%s: %v", c.Purl, err)
				message := fmt.Sprintf("License info not found for %s", c.Purl)
				componentInfo.ErrorMessage = &message
				componentInfo.ErrorCode = domain.StatusCodeToErrorCode(domain.ComponentWithoutInfo)
				clir = append(clir, componentInfo)
				continue
			}

			if len(purlLicenses) == 0 {
				s.Info("no purlLicenses data found for unversioned purl=%s.", c.Purl, c.Version)
				s.Info("no purlLicenses data found for unversioned purl=%s.", c.Purl, version)
				message := fmt.Sprintf("License info not found for %s", c.Purl)
				componentInfo.ErrorMessage = &message
				componentInfo.ErrorCode = domain.StatusCodeToErrorCode(domain.ComponentWithoutInfo)
				clir = append(clir, componentInfo)
				continue
			}
		}

		//Retrieve all the uniques licenses ids
		dedupLicensesIDs := license.ExtractLicenseIDsFromPurlLicenses(purlLicenses)
		if len(dedupLicensesIDs) == 0 {
			s.Warnf("no license data available for purl=%s version=%s", c.Purl, version)
			message := fmt.Sprintf("License info not found for %s", c.Purl)
			componentInfo.ErrorMessage = &message
			componentInfo.ErrorCode = domain.StatusCodeToErrorCode(domain.ComponentWithoutInfo)
			clir = append(clir, componentInfo)
			continue
		}

		s.Debugf("Found %d unique license_ids from all sources for purl=%s version=%s", len(dedupLicensesIDs), c.Purl, version)

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
				if !allSpdxLicenses[l] {
					allSpdxLicenses[l] = true
					fullName := ""
					url := ""
					isSpdxApproved := false
					if lu.spdxLicenseCache != nil && licenseRecord.IsSpdx {
						if detail, ok := lu.spdxLicenseCache.GetLicenseByID(l); ok {
							fullName = detail.Name
							url = detail.DetailsURL
						}
						isSpdxApproved = true
					}
					finalLicenses = append(finalLicenses, &pb.LicenseInfo{
						Id:             l,
						FullName:       fullName,
						Url:            url,
						IsSpdxApproved: isSpdxApproved,
					})
				}
			}
		}

		// If no licenses could be processed, log and continue
		if len(finalLicenses) == 0 {
			s.Warnf("no valid licenses found after processing %d license IDs for purl=%s version=%s", len(dedupLicensesIDs), c.Purl, c.Version)
			message := fmt.Sprintf("License info not found for %s", c.Purl)
			componentInfo.ErrorMessage = &message
			componentInfo.ErrorCode = domain.StatusCodeToErrorCode(domain.ComponentWithoutInfo)
			clir = append(clir, componentInfo)
			continue
		}

		// Build statement by joining all license IDs with " AND "
		var licenseIDs []string
		for _, l := range finalLicenses {
			licenseIDs = append(licenseIDs, l.Id)
		}
		//TODO: the statement should come from the DB, it's not accurate to build everything with AND
		statement := strings.Join(licenseIDs, " AND ")

		componentInfo.Statement = statement
		componentInfo.Licenses = finalLicenses
		clir = append(clir, componentInfo)

	}

	return clir, nil
}

// GetDetails
func (lu LicenseUseCase) GetDetails(ctx context.Context, s *zap.SugaredLogger, lic dto.LicenseRequestDTO) (pb.LicenseDetails, *Error) {
	licenseRecord, err := lu.licenseDetailModel.GetLicenseByID(ctx, s, lic.ID)
	if err != nil {
		return pb.LicenseDetails{}, &Error{Status: common.StatusCode_FAILED, Code: http.StatusInternalServerError, Message: err.Error(), Error: err}
	}
	if licenseRecord.ID == 0 {
		s.Warnf("LicenseDetail not found: %s", lic.ID)
		return pb.LicenseDetails{}, &Error{Status: common.StatusCode_SUCCEEDED_WITH_WARNINGS, Code: http.StatusNotFound, Message: "LicenseDetail not found", Error: errors.New("license not found")}
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
