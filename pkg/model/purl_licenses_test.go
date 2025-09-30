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
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	"testing"
)

func TestPurlLicensesModel_GetLicensesByPurl(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	db := sqliteSetup(t)
	defer CloseDB(db)

	// Load test data
	err = loadTestSQLDataFiles(db, ctx, []string{"tests/purl_licenses.sql"})
	if err != nil {
		t.Fatalf("failed to load SQL test data: %v", err)
	}

	model := NewPurlLicensesModel(db)

	t.Run("GetExistingPurlLicenses", func(t *testing.T) {
		licenses, err := model.GetLicensesByPurlVersion(ctx, "pkg:npm/express", "4.18.2")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(licenses) != 2 {
			t.Errorf("Expected 2 licenses, got: %v", len(licenses))
		}
		if licenses[0].Purl != "pkg:npm/express" {
			t.Errorf("Expected purl 'pkg:npm/express', got: %v", licenses[0].Purl)
		}
	})

	t.Run("GetNonExistentPurlLicenses", func(t *testing.T) {
		licenses, err := model.GetLicensesByPurlVersion(ctx, "pkg:npm/nonexistent", "1.0.0")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(licenses) != 0 {
			t.Errorf("Expected 0 licenses, got: %v", len(licenses))
		}
	})

	t.Run("GetLicensesWithEmptyPurl", func(t *testing.T) {
		_, err := model.GetLicensesByPurlVersion(ctx, "", "1.0.0")
		if err == nil {
			t.Error("Expected error for empty purl, got nil")
		}
	})

	t.Run("GetLicensesWithEmptyVersion", func(t *testing.T) {
		_, err := model.GetLicensesByPurlVersion(ctx, "pkg:npm/test", "")
		if err == nil {
			t.Error("Expected error for empty version, got nil")
		}
	})
}

func TestPurlLicensesModel_GetLicensesByPurlAndSource(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	db := sqliteSetup(t)
	defer CloseDB(db)

	// Load test data
	err = loadTestSQLDataFiles(db, ctx, []string{"tests/purl_licenses.sql"})
	if err != nil {
		t.Fatalf("failed to load SQL test data: %v", err)
	}

	model := NewPurlLicensesModel(db)

	t.Run("GetExistingPurlLicensesBySource", func(t *testing.T) {
		licenses, err := model.GetLicensesByPurlVersionAndSource(ctx, "pkg:npm/express", "4.18.2", []int16{1})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(licenses) != 1 {
			t.Errorf("Expected 1 license, got: %v", len(licenses))
		}
		if licenses[0].SourceID != 1 {
			t.Errorf("Expected source_id 1, got: %v", licenses[0].SourceID)
		}
		if licenses[0].LicenseID != 5614 {
			t.Errorf("Expected license_id 5614, got: %v", licenses[0].LicenseID)
		}
	})

	t.Run("GetNonExistentSourceID", func(t *testing.T) {
		licenses, err := model.GetLicensesByPurlVersionAndSource(ctx, "pkg:npm/express", "4.18.2", []int16{999})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(licenses) != 0 {
			t.Errorf("Expected 0 licenses, got: %v", len(licenses))
		}
	})
}
