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
	"testing"

	"github.com/jmoiron/sqlx"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
)

// The fixture's pkg:npm/express, which carries two sources on 4.18.2 and one on
// 4.17.1 — enough to exercise the source filtering.
const (
	expressPurl = "pkg:npm/express"
	expressID   = "eea6a5f0-ac07-5faa-86d6-42a1f88f80bd"
)

func setupLicenseModel(t *testing.T) (context.Context, *sqlx.DB, *PurlLicensesModel) {
	t.Helper()
	if err := zlog.NewSugaredDevLogger(); err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	ctx := context.Background()
	db := sqliteSetup(t)
	if err := LoadTestSQLData(db, ctx); err != nil {
		t.Fatalf("failed to load SQL test data: %v", err)
	}
	return ctx, db, NewPurlLicensesModel(db)
}

func TestGetLicensesByPurlIDVersion(t *testing.T) {
	ctx, db, model := setupLicenseModel(t)
	defer CloseDB(db)
	defer zlog.SyncZap()

	t.Run("known release", func(t *testing.T) {
		licenses, err := model.GetLicensesByPurlIDVersion(ctx, expressID, "4.18.2")
		if err != nil {
			t.Fatalf("GetLicensesByPurlIDVersion error = %v", err)
		}
		if len(licenses) != 2 {
			t.Fatalf("got %d licences, want 2", len(licenses))
		}
		for _, l := range licenses {
			// The purl is joined back in so callers keep seeing it, even though the
			// lookup is keyed by purl_id now.
			if l.Purl != expressPurl {
				t.Errorf("purl = %q, want %q", l.Purl, expressPurl)
			}
			if l.Version != "4.18.2" {
				t.Errorf("version = %q, want 4.18.2", l.Version)
			}
			// Serving records no date; the field stays empty rather than inventing one.
			if l.Date != "" {
				t.Errorf("date = %q, want empty: serving has no date column", l.Date)
			}
		}
	})

	t.Run("unknown release or component", func(t *testing.T) {
		for name, args := range map[string][2]string{
			"unknown version":   {expressID, "9.9.9"},
			"unknown component": {"00000000-0000-0000-0000-000000000000", "4.18.2"},
		} {
			t.Run(name, func(t *testing.T) {
				licenses, err := model.GetLicensesByPurlIDVersion(ctx, args[0], args[1])
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(licenses) != 0 {
					t.Errorf("got %d licences, want none", len(licenses))
				}
			})
		}
	})

	t.Run("empty arguments are rejected", func(t *testing.T) {
		if _, err := model.GetLicensesByPurlIDVersion(ctx, "", "4.18.2"); err == nil {
			t.Error("want an error for an empty component")
		}
		if _, err := model.GetLicensesByPurlIDVersion(ctx, expressID, ""); err == nil {
			t.Error("want an error for an empty version")
		}
	})
}

// The source filter is what the service's priority resolution rests on, so it is
// pinned here even though the production export currently writes 0 everywhere.
func TestGetLicensesByPurlIDVersionAndSource(t *testing.T) {
	ctx, db, model := setupLicenseModel(t)
	defer CloseDB(db)
	defer zlog.SyncZap()

	t.Run("one source of two", func(t *testing.T) {
		licenses, err := model.GetLicensesByPurlIDVersionAndSource(ctx, expressID, "4.18.2", []int16{1})
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if len(licenses) != 1 {
			t.Fatalf("got %d licences, want 1", len(licenses))
		}
		if licenses[0].SourceID != 1 {
			t.Errorf("source_id = %d, want 1", licenses[0].SourceID)
		}
	})

	t.Run("both sources", func(t *testing.T) {
		licenses, err := model.GetLicensesByPurlIDVersionAndSource(ctx, expressID, "4.18.2", []int16{1, 2})
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if len(licenses) != 2 {
			t.Errorf("got %d licences, want 2", len(licenses))
		}
	})

	t.Run("a source with no rows", func(t *testing.T) {
		licenses, err := model.GetLicensesByPurlIDVersionAndSource(ctx, expressID, "4.18.2", []int16{99})
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if len(licenses) != 0 {
			t.Errorf("got %d licences, want none", len(licenses))
		}
	})

	t.Run("no sources is rejected", func(t *testing.T) {
		if _, err := model.GetLicensesByPurlIDVersionAndSource(ctx, expressID, "4.18.2", nil); err == nil {
			t.Error("want an error when no source is given")
		}
	})
}

func TestGetLicensesByPurlIDVersionsAndSource(t *testing.T) {
	ctx, db, model := setupLicenseModel(t)
	defer CloseDB(db)
	defer zlog.SyncZap()

	licenses, err := model.GetLicensesByPurlIDVersionsAndSource(ctx, expressID,
		[]string{"4.18.2", "4.17.1"}, []int16{1, 2})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(licenses) != 3 {
		t.Fatalf("got %d licences, want 3 (two on 4.18.2, one on 4.17.1)", len(licenses))
	}
	byVersion := map[string]int{}
	for _, l := range licenses {
		byVersion[l.Version]++
	}
	if byVersion["4.18.2"] != 2 || byVersion["4.17.1"] != 1 {
		t.Errorf("per-version counts = %v, want 4.18.2:2 4.17.1:1", byVersion)
	}
}

// A missing table is a query failure, not an empty result.
func TestPurlLicensesReportsQueryFailure(t *testing.T) {
	ctx, db, model := setupLicenseModel(t)
	defer zlog.SyncZap()
	if _, err := db.Exec("DROP TABLE component_version_license;"); err != nil {
		t.Fatalf("failed to drop the table: %v", err)
	}
	defer CloseDB(db)

	if _, err := model.GetLicensesByPurlIDVersion(ctx, expressID, "4.18.2"); err == nil {
		t.Error("want an error once the table is gone")
	}
}

// serdeID is the fixture's pkg:cargo/serde: a component whose only licence is the
// package-level one, as it is for the 8.9 million components the export records in
// component_license and not in component_version_license.
const serdeID = "3d9c7a14-2b6e-5f08-9c41-7ae0d5b12f63"

func TestGetPackageLicensesByPurlID(t *testing.T) {
	ctx, db, model := setupLicenseModel(t)
	defer CloseDB(db)
	defer zlog.SyncZap()

	t.Run("a component with no per-version licence", func(t *testing.T) {
		licenses, err := model.GetPackageLicensesByPurlID(ctx, serdeID)
		if err != nil {
			t.Fatalf("GetPackageLicensesByPurlID error = %v", err)
		}
		want := []int32{109, 552}
		if len(licenses) != len(want) {
			t.Fatalf("got %d licences, want %d", len(licenses), len(want))
		}
		for i, l := range licenses {
			if l.LicenseID != want[i] {
				t.Errorf("licence[%d] = %d, want %d", i, l.LicenseID, want[i])
			}
			// The licence is the component's, not a release's. Reporting a version
			// here would invent an attribution the table does not make.
			if l.Version != "" {
				t.Errorf("version = %q, want empty", l.Version)
			}
			if l.SourceID != 0 {
				t.Errorf("source = %d, want 0 (declared)", l.SourceID)
			}
			if l.Purl != "pkg:cargo/serde" {
				t.Errorf("purl = %q, want pkg:cargo/serde", l.Purl)
			}
		}
	})

	t.Run("an unknown component returns nothing", func(t *testing.T) {
		licenses, err := model.GetPackageLicensesByPurlID(ctx, "00000000-0000-0000-0000-000000000000")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(licenses) != 0 {
			t.Errorf("got %d licences, want none", len(licenses))
		}
	})

	t.Run("an empty component is rejected", func(t *testing.T) {
		if _, err := model.GetPackageLicensesByPurlID(ctx, ""); err == nil {
			t.Error("want an error for an empty purl_id")
		}
	})
}

// A missing table is a query failure, not an empty result: a broken database must
// not be mistaken for a component with no licence on record.
// TestGetPackageLicensesWithoutTheTable pins the export without component_license
// (the 2026-10-16 build dropped it): the fallback finds nothing rather than failing,
// so callers answer "no licence" instead of logging a Catalog Error per lookup.
func TestGetPackageLicensesWithoutTheTable(t *testing.T) {
	ctx, db, model := setupLicenseModel(t)
	defer CloseDB(db)
	defer zlog.SyncZap()
	if _, err := db.Exec("DROP TABLE component_license;"); err != nil {
		t.Fatalf("failed to drop the table: %v", err)
	}
	licenses, err := model.GetPackageLicensesByPurlID(ctx, serdeID)
	if err != nil {
		t.Errorf("want no error once the table is gone, got %v", err)
	}
	if len(licenses) != 0 {
		t.Errorf("want no licences once the table is gone, got %d", len(licenses))
	}
}

// TestGetPackageLicensesReportsQueryFailure pins that a table which exists but
// cannot be queried is still an error, not mistaken for an absent one.
func TestGetPackageLicensesReportsQueryFailure(t *testing.T) {
	ctx, db, model := setupLicenseModel(t)
	defer CloseDB(db)
	defer zlog.SyncZap()
	if _, err := db.Exec("ALTER TABLE component_license RENAME COLUMN license_id TO renamed;"); err != nil {
		t.Fatalf("failed to break the table: %v", err)
	}
	if _, err := model.GetPackageLicensesByPurlID(ctx, serdeID); err == nil {
		t.Error("want an error when the table exists but the query fails")
	}
}
