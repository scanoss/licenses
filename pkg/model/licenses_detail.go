// SPDX-LicenseDetail-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2018-2022 SCANOSS.COM
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public LicenseDetail as published by
 * the Free Software Foundation, either version 2 of the LicenseDetail, or
 * (at your option) any later version.
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public LicenseDetail for more details.
 * You should have received a copy of the GNU General Public LicenseDetail
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

// Handle all interaction with the licenses table

package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

type LicenseDetailModelInterface interface {
	GetLicenseByID(ctx context.Context, s *zap.SugaredLogger, id string) (LicenseDetail, error)
}

type LicenseModel struct {
	db *sqlx.DB
}
type LicenseDetail struct {
	LicenseID             string         `json:"licenseId" db:"id"`
	Type                  string         `json:"type" db:"type"`
	Reference             string         `json:"reference" db:"reference"`
	IsDeprecatedLicenseID bool           `json:"isDeprecatedLicenseId" db:"isdeprecatedlicenseid"`
	DetailsURL            string         `json:"detailsUrl" db:"detailsurl"`
	ReferenceNumber       int            `json:"referenceNumber" db:"referencenumber"`
	Name                  string         `json:"name" db:"name"`
	SeeAlso               pq.StringArray `json:"seeAlso" db:"seealso"`
	IsOsiApproved         bool           `json:"isOsiApproved" db:"isosiapproved"`
	IsFsfLibre            bool           `json:"isFsfLibre" db:"-"`
}

// NewLicenseDetailModel create a new instance of the LicenseDetail Model.
func NewLicenseDetailModel(db *sqlx.DB) *LicenseModel {
	return &LicenseModel{db: db}
}

// GetLicenseByID retrieves license data by the given row ID.
//
// The mining table was `spdx_license_data` with all-lowercase, run-together
// column names. Serving keeps the same 825 rows under the mixed-case
// "SPDX_license_data", with the columns renamed and the identifier now SPDX_id,
// so the projection aliases them back to what LicenseDetail expects.
func (m *LicenseModel) GetLicenseByID(ctx context.Context, s *zap.SugaredLogger, licenseID string) (LicenseDetail, error) {
	licenseIDToUpper := strings.ToUpper(licenseID)
	var license LicenseDetail
	err := m.db.GetContext(ctx, &license, m.db.Rebind(
		`SELECT "SPDX_id"        AS id,
			COALESCE(type, '')       AS type,
			COALESCE(reference, '')  AS reference,
			is_deprecated            AS isdeprecatedlicenseid,
			COALESCE(details_url,'') AS detailsurl,
			reference_number         AS referencenumber,
			COALESCE(name, '')       AS name,
			-- pq.StringArray parses the Postgres array literal serving stores here;
			-- the fallback must be the empty array, not the empty string, which it
			-- cannot parse.
			COALESCE(see_also, '{}') AS seealso,
			is_osi_approved          AS isosiapproved
		FROM "SPDX_license_data" WHERE UPPER("SPDX_id") = ?`), licenseIDToUpper)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.Errorf("Error: Failed to query license table for %v: %#v", licenseIDToUpper, err)
		return LicenseDetail{}, fmt.Errorf("failed to query the license table: %w", err)
	}
	return license, nil
}
