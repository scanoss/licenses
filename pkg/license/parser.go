package license

import (
	"github.com/github/go-spdx/v2/spdxexp"
	"strings"
)

// ParseLicenseExpression parses SPDX license expressions and returns individual licenses
func ParseLicenseExpression(licenseID string) ([]string, error) {
	// Try SPDX expression parsing first
	if strings.Contains(licenseID, " AND ") || strings.Contains(licenseID, " OR ") || strings.Contains(licenseID, "(") {
		licenses, err := spdxexp.ExtractLicenses(licenseID)
		if err == nil && len(licenses) > 0 {
			return licenses, nil
		}
	}

	// Fallback to simple string splitting for legacy format
	spdxIDs := strings.Split(licenseID, "/")
	var result []string
	for _, id := range spdxIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result, nil
}
