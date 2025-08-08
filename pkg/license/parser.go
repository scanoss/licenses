package license

import (
	"github.com/github/go-spdx/v2/spdxexp"
	"strings"
)

// ParseLicenseExpression parses SPDX license expressions and returns individual licenses
func ParseLicenseExpression(license string) ([]string, error) {
	// Try SPDX expression parsing first
	if strings.Contains(license, " AND ") || strings.Contains(license, " OR ") || strings.Contains(license, "(") {
		licenses, err := spdxexp.ExtractLicenses(license)
		if err == nil && len(licenses) > 0 {
			return licenses, nil
		}
	}

	// Fallback to simple string splitting for legacy formats
	// Handle semicolon separation (e.g., "GNU GPL v2;GNU GPL v3;GNU LGPL v2.1;GNU LGPL v3")
	var spdxIDs []string
	if strings.Contains(license, ";") {
		spdxIDs = strings.Split(license, ";")
	} else {
		// Handle forward slash separation (existing legacy format)
		spdxIDs = strings.Split(license, "/")
	}

	var result []string
	for _, id := range spdxIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result, nil
}
