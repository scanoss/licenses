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
	"strings"

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

func (m *PurlLicensesModel) GetLicensesByPurlVersionAndSource(ctx context.Context, purl, version string, sourceID []int16) ([]PurlLicense, error) {
	s := ctxzap.Extract(ctx).Sugar()

	if len(purl) == 0 || len(version) == 0 {
		s.Error("Please specify a valid purl and version to query")
		return nil, errors.New("please specify a valid purl and version to query")
	}

	if len(sourceID) == 0 {
		s.Error("Please specify at least one source_id to query")
		return nil, errors.New("please specify at least one source_id to query")
	}

	// Build IN clause dynamically for multiple source_ids
	placeholders := make([]string, len(sourceID))
	args := []interface{}{purl, version}

	for i, id := range sourceID {
		placeholders[i] = fmt.Sprintf("$%d", i+3) // $3, $4, $5, etc.
		args = append(args, id)
	}

	query := fmt.Sprintf(
		"SELECT purl, version, date, source_id, license_id FROM purl_licenses WHERE purl = $1 AND version = $2 AND source_id IN (%s)",
		strings.Join(placeholders, ","),
	)

	var purlLicenses []PurlLicense
	err := m.db.SelectContext(ctx, &purlLicenses, query, args...)
	if err != nil {
		s.Errorf("Failed to query purl_licenses table for purl %v, version %v, source_ids %v: %v", purl, version, sourceID, err)
		return nil, fmt.Errorf("failed to query the purl_licenses table: %v", err)
	}

	s.Debugf("Found %v results for purl %v, version %v, source_ids %v", len(purlLicenses), purl, version, sourceID)
	return purlLicenses, nil
}

// GetLicensesByUnversionedPurlAndSource retrieves license data from unversioned purl with source filtering
func (m *PurlLicensesModel) GetLicensesByUnversionedPurlAndSource(ctx context.Context, purl string, sourceID []int16) ([]PurlLicense, error) {
	s := ctxzap.Extract(ctx).Sugar()

	if len(purl) == 0 {
		s.Error("Please specify a valid purl to query")
		return nil, errors.New("please specify a valid purl to query")
	}

	if len(sourceID) == 0 {
		s.Error("Please specify at least one source_id to query")
		return nil, errors.New("please specify at least one source_id to query")
	}

	// Build IN clause dynamically for multiple source_ids
	placeholders := make([]string, len(sourceID))
	args := []interface{}{purl}

	for i, id := range sourceID {
		placeholders[i] = fmt.Sprintf("$%d", i+2) // $2, $3, $4, etc.
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		SELECT purl, version, date, source_id, license_id 
		FROM purl_licenses 
		WHERE purl = $1 AND (version = '' OR version IS NULL) AND source_id IN (%s)
		ORDER BY source_id, license_id`, strings.Join(placeholders, ","))

	var purlLicenses []PurlLicense
	err := m.db.SelectContext(ctx, &purlLicenses, query, args...)
	if err != nil {
		s.Errorf("Failed to query unversioned license data for purl %v, source_ids %v: %v", purl, sourceID, err)
		return nil, fmt.Errorf("failed to query unversioned license data: %v", err)
	}

	s.Debugf("Found %v unversioned license results for purl %v, source_ids %v", len(purlLicenses), purl, sourceID)
	return purlLicenses, nil
}
