// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2018-2025 SCANOSS.COM
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

package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	gomodels "github.com/scanoss/go-models/pkg/models"
)

// LicenseRegistryModel reads the licence catalogue and the SPDX registry.
//
// It replaces two go-models calls that do not work against the serving schema:
// both query the mining `licenses` and `spdx_license_data` tables, which serving
// renames to `license` and "SPDX_license_data" with different column names. The
// go-models result types are reused so nothing downstream changes.
type LicenseRegistryModel struct {
	db *sqlx.DB
}

func NewLicenseRegistryModel(db *sqlx.DB) *LicenseRegistryModel {
	return &LicenseRegistryModel{db: db}
}

// GetLicenseByID reads one catalogue entry.
//
// Mining's `licenses` carried an `is_spdx` flag; serving has no such column, so
// it is derived the way the rest of the serving schema expresses it: a licence is
// SPDX if its identifier appears in the SPDX registry.
func (m *LicenseRegistryModel) GetLicenseByID(ctx context.Context, licenseID int32) (gomodels.License, error) {
	var license gomodels.License
	err := m.db.GetContext(ctx, &license, m.db.Rebind(
		`SELECT l.id                      AS id,
			COALESCE(l.long_name, '')         AS license_name,
			COALESCE(l."SPDX", '')            AS spdx_id,
			CASE WHEN sld."SPDX_id" IS NOT NULL THEN 1 ELSE 0 END AS is_spdx
		FROM license l
		LEFT JOIN "SPDX_license_data" sld ON sld."SPDX_id" = l."SPDX"
		WHERE l.id = ?`), licenseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return gomodels.License{}, fmt.Errorf("license %d not found", licenseID)
		}
		return gomodels.License{}, fmt.Errorf("failed to query the license table: %w", err)
	}
	return license, nil
}

// GetAllSPDXLicenseDetails reads the whole SPDX registry, which the cache holds
// in memory. see_also is a Postgres array literal in a text column, which the
// go-models SeeAlsoArray already parses; the fallback is the empty array rather
// than the empty string so that parse never sees a value it cannot read.
func (m *LicenseRegistryModel) GetAllSPDXLicenseDetails(ctx context.Context) ([]gomodels.SPDXLicenseDetail, error) {
	var details []gomodels.SPDXLicenseDetail
	err := m.db.SelectContext(ctx, &details,
		`SELECT "SPDX_id"                  AS id,
			COALESCE(type, '')                 AS type,
			COALESCE(reference, '')            AS reference,
			is_deprecated                      AS isdeprecatedlicenseid,
			COALESCE(details_url, '')          AS detailsurl,
			CAST(reference_number AS VARCHAR)  AS referencenumber,
			COALESCE(name, '')                 AS name,
			COALESCE(see_also, '{}')           AS seealso,
			is_osi_approved                    AS isosiapproved
		FROM "SPDX_license_data"`)
	if err != nil {
		return nil, fmt.Errorf("failed to query the SPDX license registry: %w", err)
	}
	return details, nil
}
