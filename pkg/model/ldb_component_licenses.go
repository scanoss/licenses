// SPDX-LicenseDetail-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2025 SCANOSS.COM
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

package models

import (
	"context"
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
)

type LDBComponentLicensesModel struct {
	db *sqlx.DB
}

type LDBComponentLicense struct {
	PurlMD5   string         `db:"purl_md5"`
	Source    string         `db:"source"`
	LicenseID int32          `db:"license_id"`
	License   sql.NullString `db:"license"`
}

/*
// AllURL represents a row on the AllURL table
type AllURL struct {
	Component string `db:"component"`
	Version   string `db:"version"`
	SemVer    string `db:"semver"`
	LicenseDetail   string `db:"license"`
	LicenseID int32  `db:"license_id"`
	IsSpdx    bool   `db:"is_spdx"`
	PurlName  string `db:"purl_name"`
	MineID    int32  `db:"mine_id"`
	URL       string `db:"-"` // Computed field, not from database
}

*/

// NewLDBComponentLicensesModel create a new instance of the NewLDBComponentLicensesModel Model.
func NewLDBComponentLicensesModel(db *sqlx.DB) *LDBComponentLicensesModel {
	return &LDBComponentLicensesModel{db: db}
}

func (lcl *LDBComponentLicensesModel) GetLicensesByPurlMD5(ctx context.Context, purlMD5 string) ([]LDBComponentLicense, error) {
	s := ctxzap.Extract(ctx).Sugar()

	if len(purlMD5) == 0 {
		s.Error("Please specify a valid Purl MD5 to query")
		return nil, errors.New("please specify a valid purlMD5 to query")
	}

	query := "SELECT lcl.purl_md5, lcl.source, lcl.license_id, l.license_name as license" +
		" FROM ldb_component_licenses lcl" +
		" LEFT JOIN licenses l ON lcl.license_id = l.id" +
		" WHERE purl_md5 = $1"

	var componentLicenses []LDBComponentLicense
	err := lcl.db.SelectContext(ctx, &componentLicenses, query, purlMD5)
	if err != nil {
		s.Errorf("Failed to query ldb_component_licenses table for %v: %v", purlMD5, err)
		return nil, fmt.Errorf("failed to query the ldb_component_licenses table: %v", err)
	}

	s.Debugf("Found %v results for %v", len(componentLicenses), purlMD5)
	return componentLicenses, nil
}

// CalculateMD5FromPurlVersion generates MD5 hash from component and version
func (lcl *LDBComponentLicensesModel) CalculateMD5FromPurlVersion(purl, version string) string {
	purlString := fmt.Sprintf("%s@%s", purl, version)
	hash := md5.Sum([]byte(purlString))
	return fmt.Sprintf("%x", hash)
}
