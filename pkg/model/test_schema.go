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

// The serving schema, used to build the database the tests run against.
//
// The fixtures under tests/ used to carry their own CREATE TABLE statements, each
// describing the table as that fixture's author imagined it - the osadl fixture
// declared columns the production table does not have. Keeping the schema here
// means every test is built from one description, and a query naming a column the
// real database does not have fails in CI rather than in production.
//
// This is a transcription of the serving schema, taken from the production DuckDB
// file with `SELECT sql FROM duckdb_tables()`. The types are the SQLite
// equivalents, since the tests run on SQLite: DuckDB UUID becomes TEXT, which the
// queries already tolerate because they cast it.
//
// Keep this in step with the real database. It is a transcription, so it can drift;
// when a query fails in production but passes here, suspect this file first.

package models

// testSchemaDDL is the schema every DB-backed test is built from.
const testSchemaDDL = `
CREATE TABLE component (purl TEXT, vendor TEXT, component TEXT, source_purl_id TEXT, purl_id TEXT);
CREATE INDEX idx_component_purl ON component (purl);
CREATE INDEX idx_component_purl_id ON component (purl_id);

CREATE TABLE component_version (purl_id TEXT, version TEXT, version_name TEXT, version_semver TEXT);
CREATE INDEX idx_component_version_purl_id ON component_version (purl_id);

-- Licences per release, keyed by (purl_id, version) where mining keyed them by
-- the purl string. source records provenance; the serving export currently
-- writes 0 on every row.
CREATE TABLE component_version_license (purl_id TEXT, version TEXT, license_id INTEGER, source INTEGER);
CREATE INDEX idx_component_version_license ON component_version_license (purl_id, version);

-- The licence catalogue, mining's "licenses". There is no is_spdx flag: a licence
-- is SPDX when its identifier appears in SPDX_license_data.
CREATE TABLE license (id INTEGER, long_name TEXT, SPDX TEXT, content TEXT);
CREATE INDEX idx_license_id ON license (id);

CREATE TABLE SPDX_license_data (
    SPDX_id TEXT, type TEXT, reference TEXT, is_deprecated BOOLEAN,
    details_url TEXT, reference_number INTEGER, name TEXT,
    is_osi_approved BOOLEAN, see_also TEXT
);
CREATE INDEX idx_spdx_license_data_id ON SPDX_license_data (SPDX_id);

-- OSADL. Mining kept the three compatibility lists as arrays on the row; serving
-- normalises each into its own table.
CREATE TABLE OSADL (id TEXT, copyleft TEXT, patent_hints BOOLEAN);
CREATE INDEX idx_osadl_id ON OSADL (id);
CREATE TABLE OSADL_compat (id TEXT, compat_id TEXT);
CREATE TABLE OSADL_incompat (id TEXT, incompat_id TEXT);
CREATE TABLE OSADL_depending (id TEXT, depend_id TEXT);
`
