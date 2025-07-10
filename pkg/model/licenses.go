// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2018-2022 SCANOSS.COM
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

type LicenseModelInterface interface {
	GetLicenseByID(id string) (License, error)
}

type LicenseModel struct {
	ctx context.Context
	s   *zap.SugaredLogger
	db  *sqlx.DB
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

type License struct {
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

// NewLicenseModel create a new instance of the License Model.
func NewLicenseModel(ctx context.Context, s *zap.SugaredLogger, db *sqlx.DB) *LicenseModel {
	return &LicenseModel{ctx: ctx, s: s, db: db}
}

// GetLicenseByID retrieves license data by the given row ID.
func (m *LicenseModel) GetLicenseByID(licenseId string) (License, error) {
	conn, err := NewConn(m.ctx, m.db)
	if err != nil {
		return License{}, err
	}
	licenseIDToUpper := strings.ToUpper(licenseId)
	var license License
	err = conn.QueryRowxContext(m.ctx,
		"SELECT * FROM licenses WHERE UPPER(license_id) = $1", licenseIDToUpper).StructScan(&license)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		m.s.Errorf("Error: Failed to query license table for %v: %#v", licenseIDToUpper, err)
		return License{}, fmt.Errorf("failed to query the license table: %v", err)
	}
	return license, nil
}
