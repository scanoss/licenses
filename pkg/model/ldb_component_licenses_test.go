// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2018-2023 SCANOSS.COM
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

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
)

func TestGetLicensesByPurlMD5_Success(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()
	db := sqliteSetup(t)
	defer CloseDB(db)
	conn := sqliteConn(t, ctx, db)
	defer CloseConn(conn)

	// Load test data
	err = loadTestSQLDataFiles(db, ctx, conn, []string{"tests/licenses.sql", "tests/ldb_component_licenses.sql"})
	if err != nil {
		t.Fatalf("failed to load SQL test data: %v", err)
	}

	ldbModel := NewLDBComponentLicensesModel(db)

	tests := []struct {
		name           string
		purlMD5        string
		expectedCount  int
		expectedSource string
		expectErr      bool
	}{
		{
			name:           "Single license component",
			purlMD5:        "abc123def456789",
			expectedCount:  1,
			expectedSource: "github.com/example/repo",
			expectErr:      false,
		},
		{
			name:          "Multiple licenses component",
			purlMD5:       "xyz789abc123456",
			expectedCount: 2,
			expectErr:     false,
		},
		{
			name:          "Dual license same source",
			purlMD5:       "dual456license",
			expectedCount: 2,
			expectErr:     false,
		},
		{
			name:          "Non-existent purl MD5",
			purlMD5:       "nonexistent123",
			expectedCount: 0,
			expectErr:     false,
		},
		{
			name:          "Component with orphaned license (LEFT JOIN test)",
			purlMD5:       "orphan456license",
			expectedCount: 1,
			expectErr:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			licenses, err := ldbModel.GetLicensesByPurlMD5(ctx, test.purlMD5)

			if test.expectErr {
				if err == nil {
					t.Errorf("GetLicensesByPurlMD5() error = nil, wantErr %v", test.expectErr)
				}
				return
			}

			if err != nil {
				t.Errorf("GetLicensesByPurlMD5() unexpected error = %v", err)
				return
			}

			if len(licenses) != test.expectedCount {
				t.Errorf("GetLicensesByPurlMD5() returned %d licenses, expected %d", len(licenses), test.expectedCount)
				return
			}

			if test.expectedCount > 0 {
				// Verify first license structure
				license := licenses[0]
				if license.PurlMD5 != test.purlMD5 {
					t.Errorf("GetLicensesByPurlMD5() PurlMD5 = %v, expected %v", license.PurlMD5, test.purlMD5)
				}

				if license.LicenseID == 0 {
					t.Errorf("GetLicensesByPurlMD5() LicenseID should not be 0")
				}

				// For orphaned licenses (LEFT JOIN with NULL), License.Valid will be false
				if test.purlMD5 != "orphan456license" && (!license.License.Valid || license.License.String == "") {
					t.Errorf("GetLicensesByPurlMD5() License name should not be empty for valid licenses")
				}

				if test.expectedSource != "" && license.Source != test.expectedSource {
					t.Errorf("GetLicensesByPurlMD5() Source = %v, expected %v", license.Source, test.expectedSource)
				}

				// Special validation for orphaned license (LEFT JOIN with NULL)
				if test.purlMD5 == "orphan456license" {
					if license.License.Valid {
						t.Errorf("GetLicensesByPurlMD5() Expected NULL license for orphaned license, got %v", license.License.String)
					}
					if license.LicenseID != 99999 {
						t.Errorf("GetLicensesByPurlMD5() Expected license_id 99999 for orphaned license, got %d", license.LicenseID)
					}
				}
			}

			s.Debugf("Test %s: Found %d licenses for purlMD5 %s", test.name, len(licenses), test.purlMD5)
		})
	}
}

func TestGetLicensesByPurlMD5_EmptyInput(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	db := sqliteSetup(t)
	defer CloseDB(db)
	conn := sqliteConn(t, ctx, db)
	defer CloseConn(conn)

	// Load test data
	err = loadTestSQLDataFiles(db, ctx, conn, []string{"tests/licenses.sql", "tests/ldb_component_licenses.sql"})
	if err != nil {
		t.Fatalf("failed to load SQL test data: %v", err)
	}

	ldbModel := NewLDBComponentLicensesModel(db)

	tests := []struct {
		name    string
		purlMD5 string
	}{
		{
			name:    "Empty string",
			purlMD5: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			licenses, err := ldbModel.GetLicensesByPurlMD5(ctx, test.purlMD5)

			if err == nil {
				t.Errorf("GetLicensesByPurlMD5() with empty input should return an error")
				return
			}

			if licenses != nil {
				t.Errorf("GetLicensesByPurlMD5() with empty input should return nil licenses")
			}

			expectedErrMsg := "please specify a valid purlMD5 to query"
			if err.Error() != expectedErrMsg {
				t.Errorf("GetLicensesByPurlMD5() error = %v, expected %v", err.Error(), expectedErrMsg)
			}
		})
	}
}

func TestGetLicensesByPurlMD5_DatabaseError(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	db := sqliteSetup(t)
	defer CloseDB(db)

	// Don't load test data to cause a database error (missing table)
	ldbModel := NewLDBComponentLicensesModel(db)

	_, err = ldbModel.GetLicensesByPurlMD5(ctx, "test123")
	if err == nil {
		t.Errorf("GetLicensesByPurlMD5() should return error when table doesn't exist")
	}
}
