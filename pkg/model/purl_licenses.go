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
	"errors"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
)

// PurlLicensesModel reads the licences recorded against a component's releases.
//
// The mining table was `purl_licenses(purl, version, date, source_id, license_id)`,
// keyed by the purl string. Serving keys them by the component's purl_id in
// `component_version_license(purl_id, version, license_id, source)`, so every
// lookup takes a resolved purl_id rather than a purl.
//
// Two columns behave differently and the callers should know:
//   - there is no `date`; the field is kept on the struct (nothing reads it) and
//     comes back empty;
//   - `source` is currently 0 on every row of the serving export, so the
//     source-priority resolution the service performs always settles on the
//     first configured source. The filtering is kept as-is: it is correct code
//     against a column the export is expected to populate.
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

// licenseSelect is the projection every lookup shares. `date` is selected as a
// literal because the serving schema does not record one.
const licenseSelect = `SELECT c.purl AS purl, cvl.version AS version, '' AS date,
	cvl.source     AS source_id,
	cvl.license_id AS license_id
FROM component_version_license cvl
INNER JOIN component c ON c.purl_id = cvl.purl_id
WHERE cvl.purl_id = ?`

// GetLicensesByPurlIDVersion returns every licence recorded for one release.
func (m *PurlLicensesModel) GetLicensesByPurlIDVersion(ctx context.Context, purlID, version string) ([]PurlLicense, error) {
	s := ctxzap.Extract(ctx).Sugar()
	if len(purlID) == 0 || len(version) == 0 {
		s.Error("Please specify a valid component and version to query")
		return nil, errors.New("please specify a valid component and version to query")
	}
	var purlLicenses []PurlLicense
	err := m.db.SelectContext(ctx, &purlLicenses,
		m.db.Rebind(licenseSelect+" AND cvl.version = ?"), purlID, version)
	if err != nil {
		s.Errorf("Failed to query component licences for %v, version %v: %v", purlID, version, err)
		return nil, fmt.Errorf("failed to query the component licences: %w", err)
	}
	s.Debugf("Found %v results for %v, version %v", len(purlLicenses), purlID, version)
	return purlLicenses, nil
}

// GetLicensesByPurlIDVersionAndSource returns one release's licences, restricted
// to the given sources.
func (m *PurlLicensesModel) GetLicensesByPurlIDVersionAndSource(ctx context.Context, purlID, version string, sourceID []int16) ([]PurlLicense, error) {
	s := ctxzap.Extract(ctx).Sugar()
	if len(purlID) == 0 || len(version) == 0 {
		s.Error("Please specify a valid component and version to query")
		return nil, errors.New("please specify a valid component and version to query")
	}
	if len(sourceID) == 0 {
		s.Error("Please specify at least one source_id to query")
		return nil, errors.New("please specify at least one source_id to query")
	}
	query, args, err := sqlx.In(licenseSelect+" AND cvl.version = ? AND cvl.source IN (?)", purlID, version, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to build the component licences query: %w", err)
	}
	var purlLicenses []PurlLicense
	if err = m.db.SelectContext(ctx, &purlLicenses, m.db.Rebind(query), args...); err != nil {
		s.Errorf("Failed to query component licences for %v, version %v, source_ids %v: %v", purlID, version, sourceID, err)
		return nil, fmt.Errorf("failed to query the component licences: %w", err)
	}
	s.Debugf("Found %v results for %v, version %v, source_ids %v", len(purlLicenses), purlID, version, sourceID)
	return purlLicenses, nil
}

// GetLicensesByPurlIDVersionsAndSource returns the licences of several releases
// of one component, restricted to the given sources. The package-level aggregate
// (the empty version) is excluded, as it was in mining.
func (m *PurlLicensesModel) GetLicensesByPurlIDVersionsAndSource(ctx context.Context, purlID string, versions []string, sourceID []int16) ([]PurlLicense, error) {
	s := ctxzap.Extract(ctx).Sugar()
	if len(purlID) == 0 {
		s.Error("Please specify a valid component to query")
		return nil, errors.New("please specify a valid component to query")
	}
	if len(versions) == 0 {
		s.Error("Please specify at least one version to query")
		return nil, errors.New("please specify at least one version to query")
	}
	if len(sourceID) == 0 {
		s.Error("Please specify at least one source_id to query")
		return nil, errors.New("please specify at least one source_id to query")
	}
	query, args, err := sqlx.In(
		licenseSelect+" AND cvl.version IN (?) AND cvl.source IN (?) AND cvl.version <> ''",
		purlID, versions, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to build the component licences query: %w", err)
	}
	var purlLicenses []PurlLicense
	if err = m.db.SelectContext(ctx, &purlLicenses, m.db.Rebind(query), args...); err != nil {
		s.Errorf("Failed to query component licences for %v, %d versions, source_ids %v: %v", purlID, len(versions), sourceID, err)
		return nil, fmt.Errorf("failed to query the component licences: %w", err)
	}
	s.Debugf("Found %v results for %v across %d versions, source_ids %v", len(purlLicenses), purlID, len(versions), sourceID)
	return purlLicenses, nil
}

// GetLicensesByUnversionedPurlIDAndSource returns the component-level licences:
// the rows serving stores against the empty version, which is the package-level
// aggregate rather than any particular release.
func (m *PurlLicensesModel) GetLicensesByUnversionedPurlIDAndSource(ctx context.Context, purlID string, sourceID []int16) ([]PurlLicense, error) {
	s := ctxzap.Extract(ctx).Sugar()
	if len(purlID) == 0 {
		s.Error("Please specify a valid component to query")
		return nil, errors.New("please specify a valid component to query")
	}
	if len(sourceID) == 0 {
		s.Error("Please specify at least one source_id to query")
		return nil, errors.New("please specify at least one source_id to query")
	}
	query, args, err := sqlx.In(
		licenseSelect+" AND (cvl.version = '' OR cvl.version IS NULL) AND cvl.source IN (?)"+
			" ORDER BY cvl.source, cvl.license_id",
		purlID, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to build the component licences query: %w", err)
	}
	var purlLicenses []PurlLicense
	if err = m.db.SelectContext(ctx, &purlLicenses, m.db.Rebind(query), args...); err != nil {
		s.Errorf("Failed to query unversioned licences for %v, source_ids %v: %v", purlID, sourceID, err)
		return nil, fmt.Errorf("failed to query unversioned licence data: %w", err)
	}
	s.Debugf("Found %v unversioned licence results for %v, source_ids %v", len(purlLicenses), purlID, sourceID)
	return purlLicenses, nil
}
