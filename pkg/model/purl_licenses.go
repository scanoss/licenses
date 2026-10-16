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
	"strings"
	"sync"

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

	// packageTable records whether `component_license` exists, once known. The
	// export dropped it in the 2026-10-16 build; without this every lookup that
	// reached the fallback logged a Catalog Error.
	packageTableMu    sync.Mutex
	packageTableKnown bool
	packageTable      bool
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

// packageLicenseSelect reads the component-level licences from `component_license`,
// the table that records a component's licence without attributing it to a release.
//
// Version and source are selected as literals because the table has neither: the
// row is the component's licence, not a release's, and everything the export
// records as a licence is declared (source 0 - `component_version_license.source`
// is 0 on all 205 million of its rows).
//
// This is a fallback, not a source in its own right. `component_version_license`
// covers 22.8 million components and `component_license` covers all 31.8 million,
// so 8.9 million components have a licence here and nowhere else; without this they
// answer "no licence found" when one is on record.
const packageLicenseSelect = `SELECT c.purl AS purl, '' AS version, '' AS date,
	0            AS source_id,
	cl.license_id AS license_id
FROM component_license cl
INNER JOIN component c ON c.purl_id = cl.purl_id
WHERE cl.purl_id = ?
ORDER BY cl.license_id`

// GetPackageLicensesByPurlID returns a component's licences as recorded against the
// component rather than any release.
//
// Callers use it only once a version lookup has come back empty, so a component
// with per-release licences keeps reporting those, with their versions intact. The
// rows it returns carry an empty version, because the licence genuinely is not
// attributed to one - reporting the requested version would invent an attribution
// the database does not make.
func (m *PurlLicensesModel) GetPackageLicensesByPurlID(ctx context.Context, purlID string) ([]PurlLicense, error) {
	s := ctxzap.Extract(ctx).Sugar()
	if len(purlID) == 0 {
		s.Error("Please specify a valid component to query")
		return nil, errors.New("please specify a valid component to query")
	}
	if !m.hasPackageLicenseTable(ctx) {
		return nil, nil
	}
	var purlLicenses []PurlLicense
	if err := m.db.SelectContext(ctx, &purlLicenses, m.db.Rebind(packageLicenseSelect), purlID); err != nil {
		s.Errorf("Failed to query package licences for %v: %v", purlID, err)
		return nil, fmt.Errorf("failed to query the package licences: %w", err)
	}
	s.Debugf("Found %v package-level licence results for %v", len(purlLicenses), purlID)
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

// hasPackageLicenseTable reports whether the export carries `component_license`.
//
// It probes once and remembers the answer only when the probe is conclusive: a
// missing-table error means the fallback has nothing to read, while any other
// failure is treated as transient and the probe runs again next time, so a
// momentary error cannot switch the fallback off for the life of the process.
func (m *PurlLicensesModel) hasPackageLicenseTable(ctx context.Context) bool {
	m.packageTableMu.Lock()
	defer m.packageTableMu.Unlock()
	if m.packageTableKnown {
		return m.packageTable
	}
	var probe []int
	err := m.db.SelectContext(ctx, &probe, "SELECT 1 FROM component_license LIMIT 1")
	switch {
	case err == nil:
		m.packageTableKnown, m.packageTable = true, true
	case isMissingTable(err):
		ctxzap.Extract(ctx).Sugar().Warnf("component_license is not in this database; " +
			"the package-level licence fallback is disabled")
		m.packageTableKnown, m.packageTable = true, false
	default:
		return false
	}
	return m.packageTable
}

// isMissingTable recognises the missing-table error of the engines the service
// runs on: DuckDB ("Catalog Error: Table with name ... does not exist") and SQLite
// ("no such table").
func isMissingTable(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "Catalog Error: Table with name") || strings.Contains(msg, "no such table")
}
