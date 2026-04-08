package usecase

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-component-helper/componenthelper"
	comphelputils "github.com/scanoss/go-component-helper/componenthelper/utils"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	"github.com/scanoss/go-models/pkg/scanoss"
	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/licensesv2"
	"go.uber.org/zap"
	"scanoss.com/licenses/pkg/cache"
	myconfig "scanoss.com/licenses/pkg/config"
	"scanoss.com/licenses/pkg/dto"
	"scanoss.com/licenses/pkg/license"
	models "scanoss.com/licenses/pkg/model"
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

// NewLicenseUseCaseWithLicenseModel option for dependency injection (mainly for testing).
func NewLicenseUseCaseWithLicenseModel(config *myconfig.ServerConfig, licenseModel models.LicenseDetailModelInterface,
	osadlModel models.OSADLModelInterface) *LicenseUseCase {
	return &LicenseUseCase{
		config:             config,
		licenseDetailModel: licenseModel,
		osadlModel:         osadlModel,
	}
}

// GetComponentLicense retrieves license info for a single component.
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

// GetComponentsLicense retrieves license info for multiple components.
func (lu LicenseUseCase) GetComponentsLicense(ctx context.Context, componentDTOs []componenthelper.ComponentDTO) ([]*pb.ComponentLicenseInfo, *Error) {
	s := ctxzap.Extract(ctx).Sugar()
	processedComponents := componenthelper.GetComponentsVersion(componenthelper.ComponentVersionCfg{
		MaxWorkers: 5,
		DB:         lu.db,
		Ctx:        ctx,
		S:          s,
		Input:      componentDTOs,
	})
	clir := make([]*pb.ComponentLicenseInfo, 0, len(processedComponents))
	for _, c := range processedComponents {
		if c.Status.StatusCode != domain.Success && c.Status.StatusCode != domain.VersionNotFound {
			msg := c.Status.Message
			clir = append(clir, &pb.ComponentLicenseInfo{
				Purl:         c.OriginalPurl,
				Requirement:  c.OriginalRequirement,
				Version:      c.Version,
				Url:          c.URL,
				ErrorMessage: &msg,
				ErrorCode:    domain.StatusCodeToErrorCode(c.Status.StatusCode),
			})
			continue
		}
		clir = append(clir, lu.processComponentLicenses(ctx, s, c))
	}
	return clir, nil
}

func (lu LicenseUseCase) processComponentLicenses(ctx context.Context, s *zap.SugaredLogger,
	c componenthelper.Component) *pb.ComponentLicenseInfo {
	componentInfo := &pb.ComponentLicenseInfo{
		Purl:        c.OriginalPurl,
		Requirement: c.OriginalRequirement,
		Url:         c.URL,
	}

	version := c.Version
	var purlLicenses []models.PurlLicense

	// Step 1: Try to fetch licenses for the exact resolved version (e.g. "1.2.3").
	if version != "" {
		purlLicenses = lu.fetchLicensesByPurlAndVersion(ctx, s, c.Purl, version)
	}

	// Step 2: If no licenses were found for the exact version and there are known versions available,
	// query all known versions at once and pick the nearest version to the requirement that has license data.
	// If no requirement was provided, fall back to using the resolved version as the reference point.
	if len(purlLicenses) == 0 && len(c.Versions) > 0 {
		requirement := c.Requirement
		if requirement == "" {
			requirement = c.Version
		}
		nearestLicenses, nearestVersion := lu.fetchLicensesByPurlAndVersions(ctx, s, c.Purl, requirement, c.Versions)
		if len(nearestLicenses) > 0 {
			purlLicenses = nearestLicenses
			version = nearestVersion
			// When a requirement was explicitly provided, check if the nearest version actually satisfies it.
			// If it doesn't, inform the caller that the returned version doesn't meet the original constraint.
			if c.Requirement != "" {
				if !versionSatisfiesRequirement(nearestVersion, c.Requirement) {
					message := fmt.Sprintf("Version not found for requirement %s, nearest version found: %s", c.Requirement, version)
					componentInfo.ErrorMessage = &message
					componentInfo.ErrorCode = domain.StatusCodeToErrorCode(domain.VersionNotFound)
				}
			}
		}
	}

	// Step 3: Last resort — try fetching licenses from the unversioned purl entry.
	if len(purlLicenses) == 0 {
		s.Infof("no purlLicenses data found for purl=%s version=%s. Trying unversioned purl", c.Purl, version)
		purlLicenses = lu.fetchLicensesByPurl(ctx, s, c.Purl, []int16{license.SourceScancodeAttributionFiles})
		version = ""
		componentInfo.ErrorCode = domain.StatusCodeToErrorCode(domain.VersionNotFound)
		message := "Retrieving licenses for unversioned component"
		if len(purlLicenses) > 0 && c.Requirement != "" {
			message = fmt.Sprintf("Licenses for requirement %s not found, retrieving licenses for unversioned component", c.Requirement)
		}
		componentInfo.ErrorMessage = &message
	}

	componentInfo.Version = version

	if len(purlLicenses) == 0 {
		message := fmt.Sprintf("License info not found for %s", c.Purl)
		componentInfo.ErrorMessage = &message
		componentInfo.ErrorCode = domain.StatusCodeToErrorCode(domain.ComponentWithoutInfo)
		return componentInfo
	}

	// Retrieve all the unique license ids
	dedupLicensesIDs := license.ExtractLicenseIDsFromPurlLicenses(purlLicenses)
	if len(dedupLicensesIDs) == 0 {
		s.Warnf("no license data available for purl=%s version=%s", c.Purl, version)
		message := fmt.Sprintf("License info not found for %s", c.Purl)
		componentInfo.ErrorMessage = &message
		componentInfo.ErrorCode = domain.StatusCodeToErrorCode(domain.ComponentWithoutInfo)
		return componentInfo
	}

	s.Debugf("Found %d unique license_ids from all sources for purl=%s version=%s", len(dedupLicensesIDs), c.Purl, version)

	finalLicenses := lu.resolveSPDXLicenses(ctx, s, dedupLicensesIDs)

	// If no licenses could be processed, log and return
	if len(finalLicenses) == 0 {
		s.Warnf("no valid licenses found after processing %d license IDs for purl=%s version=%s", len(dedupLicensesIDs), c.Purl, c.Version)
		message := fmt.Sprintf("License info not found for %s", c.Purl)
		componentInfo.ErrorMessage = &message
		componentInfo.ErrorCode = domain.StatusCodeToErrorCode(domain.ComponentWithoutInfo)
		return componentInfo
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
	return componentInfo
}

// fetchLicensesByPurlAndVersion retrieves licenses for a specific purl and version.
func (lu LicenseUseCase) fetchLicensesByPurlAndVersion(ctx context.Context, s *zap.SugaredLogger,
	purl, version string) []models.PurlLicense {
	purlLicenses, err := lu.purlLicenseModel.GetLicensesByPurlVersionAndSource(ctx, purl, version, lu.config.Lookup.SourcePriority)
	if err != nil {
		s.Warnf("error when querying GetLicensesByPurlVersionAndSource() for purl=%s version=%s: %v", purl, version, err)
		return nil
	}
	return purlLicenses
}

// fetchLicensesByPurlAndVersions retrieves licenses across multiple versions for a purl
// and returns the licenses for the nearest version to the requirement.
func (lu LicenseUseCase) fetchLicensesByPurlAndVersions(ctx context.Context, s *zap.SugaredLogger,
	purl, requirement string, versions []string) ([]models.PurlLicense, string) {
	allVersionLicenses, err := lu.purlLicenseModel.GetLicensesByPurlVersionsAndSource(ctx, purl, versions, lu.config.Lookup.SourcePriority)
	if err != nil {
		s.Warnf("error when querying GetLicensesByPurlVersionsAndSource() for purl=%s: %v", purl, err)
		return nil, ""
	}
	if len(allVersionLicenses) == 0 {
		return nil, ""
	}

	// Group licenses by version
	licensesByVersion := make(map[string][]models.PurlLicense)
	for _, pl := range allVersionLicenses {
		licensesByVersion[pl.Version] = append(licensesByVersion[pl.Version], pl)
	}

	nearestVersion := comphelputils.FindNearestVersion(requirement, slices.Collect(maps.Keys(licensesByVersion)))
	if nearestVersion == "" {
		return nil, ""
	}

	s.Debugf("using nearest version %s (from %d available) for purl=%s", nearestVersion, len(licensesByVersion), purl)
	return licensesByVersion[nearestVersion], nearestVersion
}

// fetchLicensesByPurl retrieves licenses for an unversioned purl.
func (lu LicenseUseCase) fetchLicensesByPurl(ctx context.Context, s *zap.SugaredLogger,
	purl string, sourceID []int16) []models.PurlLicense {
	purlLicenses, err := lu.purlLicenseModel.GetLicensesByUnversionedPurlAndSource(ctx, purl, sourceID)
	if err != nil {
		s.Warnf("error when querying GetLicensesByUnversionedPurlAndSource() for purl=%s: %v", purl, err)
		return nil
	}
	return purlLicenses
}

// versionSatisfiesRequirement checks if a version satisfies a semver constraint/requirement.
func versionSatisfiesRequirement(version, requirement string) bool {
	c, err := semver.NewConstraint(requirement)
	if err != nil {
		return false
	}
	v, err := semver.NewVersion(version)
	if err != nil {
		return false
	}
	return c.Check(v)
}

func (lu LicenseUseCase) resolveSPDXLicenses(ctx context.Context, s *zap.SugaredLogger,
	dedupLicensesIDs []int32) []*pb.LicenseInfo {
	var finalLicenses []*pb.LicenseInfo
	allSpdxLicenses := make(map[string]bool)

	for _, licenseID := range dedupLicensesIDs {
		licenseRecord, err := lu.sc.Models.Licenses.GetLicenseByID(ctx, licenseID)
		if err != nil {
			s.Warnf("error getting license by ID: %d. %v", licenseID, err)
			continue
		}

		spdx, err := license.ParseLicenseExpression(licenseRecord.SPDX)
		if err != nil {
			s.Warnf("error parsing license expression for license_id %d: %s. %v", licenseID, licenseRecord.SPDX, err)
			continue
		}

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

	return finalLicenses
}

// GetDetails retrieves detailed license information.
func (lu LicenseUseCase) GetDetails(ctx context.Context, s *zap.SugaredLogger, lic dto.LicenseRequestDTO) (pb.LicenseDetails, *Error) {
	licenseRecord, err := lu.licenseDetailModel.GetLicenseByID(ctx, s, lic.ID)
	if err != nil {
		return pb.LicenseDetails{}, &Error{Status: common.StatusCode_FAILED, Code: http.StatusInternalServerError, Message: err.Error(), Error: err}
	}
	if licenseRecord.ID == 0 {
		s.Warnf("LicenseDetail not found: %s", lic.ID)
		return pb.LicenseDetails{}, &Error{
			Status: common.StatusCode_SUCCEEDED_WITH_WARNINGS,
			Code:   http.StatusNotFound, Message: "LicenseDetail not found",
			Error: errors.New("license not found")}
	}
	s.Debugf("LicenseDetail: %v", licenseRecord)

	osadl, err := lu.osadlModel.GetOSADLByLicenseID(ctx, s, licenseRecord.LicenseID)
	if err != nil {
		s.Errorf("Error getting OSADL for license: %s, err: %v\n", lic.ID, err)
	}

	s.Debugf("OSADL: %v", osadl)

	return pb.LicenseDetails{
		FullName: licenseRecord.Name,
		Spdx: &pb.SPDX{
			FullName:      licenseRecord.Name,
			Id:            licenseRecord.LicenseID,
			DetailsUrl:    licenseRecord.DetailsURL,
			ReferenceUrl:  licenseRecord.Reference,
			IsDeprecated:  licenseRecord.IsDeprecatedLicenseID,
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
