// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2026 SCANOSS.COM
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 2 of the License, or
 * (at your option) any later version.
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package license

import (
	models "scanoss.com/licenses/pkg/model"
)

// PickLicensesByPriority walks sourcePriority in order and returns the rows from licenses
// belonging to the first source that has at least one matching row. Returns nil if no source
// in sourcePriority has any rows in licenses.
func PickLicensesByPriority(licenses []models.PurlLicense, sourcePriority []int16) []models.PurlLicense {
	for _, source := range sourcePriority {
		var picked []models.PurlLicense
		for _, l := range licenses {
			if l.SourceID == source {
				picked = append(picked, l)
			}
		}
		if len(picked) > 0 {
			return picked
		}
	}
	return nil
}

// ExtractLicenseIDsFromPurlLicenses extracts all unique SPDX licenses from all license_ids.
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
