// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2025 SCANOSS.COM
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
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"github.com/jmoiron/sqlx"
)

type PurlLicensesModel struct {
	db *sqlx.DB
}

type PurlLicense struct {
	Purl      string `json:"purl" db:"purl"`
	Version   string `json:"version" db:"version"`
	Date      string `json:"date" db:"date"`
	SourceID  int16  `json:"source_id" db:"source_id"`
	LicenseID int32  `json:"license_id" db:"license_id"`
}

func NewPurlLicensesModel(db *sqlx.DB) *PurlLicensesModel {
	return &PurlLicensesModel{db: db}
}

func (m *PurlLicensesModel) GetLicensesByPurl(ctx context.Context, purl, version string) ([]PurlLicense, error) {
	s := ctxzap.Extract(ctx).Sugar()
	if len(purl) == 0 || len(version) == 0 {
		s.Error("Please specify a valid purl and version to query")
		return nil, errors.New("please specify a valid purl and version to query")
	}

	var purlLicenses []PurlLicense
	err := m.db.SelectContext(ctx, &purlLicenses,
		"SELECT purl, version, date, source_id, license_id FROM purl_licenses WHERE purl = $1 AND version IS NULL",
		purl, version)
	if err != nil {
		s.Errorf("Failed to query purl_licenses table for purl %v, version %v: %v", purl, version, err)
		return nil, fmt.Errorf("failed to query the purl_licenses table: %v", err)
	}

	s.Debugf("Found %v results for purl %v, version %v", len(purlLicenses), purl, version)
	return purlLicenses, nil
}

func (m *PurlLicensesModel) GetLicensesByPurlVersion(ctx context.Context, purl, version string) ([]PurlLicense, error) {
	s := ctxzap.Extract(ctx).Sugar()
	if len(purl) == 0 || len(version) == 0 {
		s.Error("Please specify a valid purl and version to query")
		return nil, errors.New("please specify a valid purl and version to query")
	}

	var purlLicenses []PurlLicense
	err := m.db.SelectContext(ctx, &purlLicenses,
		"SELECT purl, version, date, source_id, license_id FROM purl_licenses WHERE purl = $1 AND version = $2",
		purl, version)
	if err != nil {
		s.Errorf("Failed to query purl_licenses table for purl %v, version %v: %v", purl, version, err)
		return nil, fmt.Errorf("failed to query the purl_licenses table: %v", err)
	}

	s.Debugf("Found %v results for purl %v, version %v", len(purlLicenses), purl, version)
	return purlLicenses, nil
}

func (m *PurlLicensesModel) GetLicensesByPurlAndSource(ctx context.Context, purl, version string, sourceID int16) ([]PurlLicense, error) {
	s := ctxzap.Extract(ctx).Sugar()

	if len(purl) == 0 || len(version) == 0 {
		s.Error("Please specify a valid purl and version to query")
		return nil, errors.New("please specify a valid purl and version to query")
	}

	var purlLicenses []PurlLicense
	err := m.db.SelectContext(ctx, &purlLicenses,
		"SELECT purl, version, date, source_id, license_id FROM purl_licenses WHERE purl = $1 AND version = $2 AND source_id = $3",
		purl, version, sourceID)
	if err != nil {
		s.Errorf("Failed to query purl_licenses table for purl %v, version %v, source_id %v: %v", purl, version, sourceID, err)
		return nil, fmt.Errorf("failed to query the purl_licenses table: %v", err)
	}

	s.Debugf("Found %v results for purl %v, version %v, source_id %v", len(purlLicenses), purl, version, sourceID)
	return purlLicenses, nil
}
