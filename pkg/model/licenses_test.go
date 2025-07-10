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
	"fmt"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
)

func TestLicensesById(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()
	db := sqliteSetup(t) // Setup SQL Lite DB
	defer CloseDB(db)
	conn := sqliteConn(t, ctx, db) // Get a connection from the pool
	defer CloseConn(conn)
	err = loadTestSQLDataFiles(db, ctx, conn, []string{"tests/licenses.sql"})
	if err != nil {
		t.Fatalf("failed to load SQL test data: %v", err)
	}
	licenseModel := NewLicenseModel(ctx, s, db)

	tests := []struct {
		licenseID string
		expectErr bool
	}{
		{
			licenseID: "MIT",
			expectErr: false,
		},
		{
			licenseID: "my-license",
			expectErr: false,
		},
		{
			licenseID: "",
			expectErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.licenseID, func(t *testing.T) {
			license, err := licenseModel.GetLicenseByID(test.licenseID)
			if test.expectErr {
				if err == nil {
					t.Errorf("licenses.GetLicenseByID() error = %v, wantErr %v", err, test.expectErr)
				}
			}
			fmt.Printf("License: %#v\n", license)
		})
	}
}
