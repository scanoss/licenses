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
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"strings"
)

type LicenseDetailModelInterface interface {
	GetLicenseByID(ctx context.Context, s *zap.SugaredLogger, id string) (LicenseDetail, error)
}

type LicenseModel struct {
	db *sqlx.DB
}

type SeeAlso []string

func (s *SeeAlso) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	str, ok := value.(string)
	if !ok {
		*s = nil
		return nil
	}

	if str == "" {
		*s = nil
		return nil
	}

	var result []string
	if err := json.Unmarshal([]byte(str), &result); err != nil {
		*s = nil
		return nil
	}

	*s = SeeAlso(result)
	return nil
}

func (s SeeAlso) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}

	data, err := json.Marshal([]string(s))
	if err != nil {
		return nil, err
	}

	return string(data), nil
}

type LicenseDetail struct {
	ID                    int32   `json:"id" db:"id"`
	Reference             string  `json:"reference" db:"reference"`
	IsDeprecatedLicenseId bool    `json:"isDeprecatedLicenseId" db:"is_deprecated_license_id"`
	DetailsUrl            string  `json:"detailsUrl" db:"details_url"`
	ReferenceNumber       int     `json:"referenceNumber" db:"reference_number"`
	Name                  string  `json:"name" db:"name"`
	LicenseId             string  `json:"licenseId" db:"license_id"`
	SeeAlso               SeeAlso `json:"seeAlso" db:"see_also"`
	IsOsiApproved         bool    `json:"isOsiApproved" db:"is_osi_approved"`
	IsFsfLibre            bool    `json:"isFsfLibre" db:"is_fsf_libre"`
}

// NewLicenseDetailModel create a new instance of the LicenseDetail Model.
func NewLicenseDetailModel(db *sqlx.DB) *LicenseModel {
	return &LicenseModel{db: db}
}

// GetLicenseByID retrieves license data by the given row ID.
func (m *LicenseModel) GetLicenseByID(ctx context.Context, s *zap.SugaredLogger, licenseId string) (LicenseDetail, error) {
	conn, err := NewConn(ctx, m.db)
	if err != nil {
		return LicenseDetail{}, err
	}
	licenseIDToUpper := strings.ToUpper(licenseId)
	var license LicenseDetail
	err = conn.QueryRowxContext(ctx,
		"SELECT * FROM licenses WHERE UPPER(license_id) = $1", licenseIDToUpper).StructScan(&license)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.Errorf("Error: Failed to query license table for %v: %#v", licenseIDToUpper, err)
		return LicenseDetail{}, fmt.Errorf("failed to query the license table: %v", err)
	}
	return license, nil
}
