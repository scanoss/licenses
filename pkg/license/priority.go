package license

import (
	models "scanoss.com/licenses/pkg/model"
	"sort"
)

// Source ID constants for license detection methods
const (
	SourceComponentDeclared        = 0
	SourceInternalAttributionFiles = 3
	SourceSPDXAttributionFiles     = 6
	SourceScancodeAttributionFiles = 5
)

// Component-level source priority order
var componentSourcePriorityOrder = []int16{
	// highest priority
	SourceSPDXAttributionFiles,     // Detected from SPDX-License-Identifier tag in attribution files
	SourceInternalAttributionFiles, // Internal mechanism detected in attribution files (LICENSE, COPYING, META-INF, etc)
	SourceScancodeAttributionFiles, // Component level through scancode in attribution files
	SourceComponentDeclared,        // Component Declared
}

// GetPriorityLevel returns the priority level for a given source_id (lower index = higher priority)
func GetPriorityLevel(sourceID int16) int {
	for index, sourceInArray := range componentSourcePriorityOrder {
		if sourceInArray == sourceID {
			return index
		}
	}
	return 999 // Unknown source_id gets lowest priority
}

// SelectBestLicense selects the license with the highest priority source_id
func SelectBestLicense(licenses []models.PurlLicense) models.PurlLicense {
	if len(licenses) == 0 {
		return models.PurlLicense{}
	}
	if len(licenses) == 1 {
		return licenses[0]
	}

	// Sort by priority (lower priority value = higher priority)
	sort.Slice(licenses, func(i, j int) bool {
		return GetPriorityLevel(licenses[i].SourceID) < GetPriorityLevel(licenses[j].SourceID)
	})

	return licenses[0]
}
