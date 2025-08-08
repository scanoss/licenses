package license

import (
	models "scanoss.com/licenses/pkg/model"
)

// Source ID constants for license detection methods
const (
	SourceComponentDeclared        = int16(0)
	SourceInternalAttributionFiles = int16(3)
	SourceSPDXAttributionFiles     = int16(6)
	SourceScancodeAttributionFiles = int16(5)
)

// ExtractLicenseIDsFromPurlLicenses extracts all unique SPDX licenses from all license_ids
func ExtractLicenseIDsFromPurlLicenses(licenses []models.PurlLicense) []int32 {
	if len(licenses) == 0 {
		return []int32{}
	}

	// Collect all unique license_ids from all sources
	allLicenseIDs := make(map[int32]bool)

	for _, license := range licenses {
		allLicenseIDs[license.LicenseID] = true
	}

	// Convert map to slice
	var result []int32
	for licenseID := range allLicenseIDs {
		result = append(result, licenseID)
	}

	return result
}
