// SPDX-LicenseDetail-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2018-2022 SCANOSS.COM
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public LicenseDetail as published by
 * the Free Software Foundation, either version 2 of the LicenseDetail, or
 * (at your option) any later version.
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public LicenseDetail for more details.
 * You should have received a copy of the GNU General Public LicenseDetail
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package models

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestDbLoad(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer CloseDB(db)
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	err = loadSQLData(db, ctx, "./tests/licenses.sql")
	if err != nil {
		t.Errorf("failed to load SQL test data: %v", err)
	}
	err = LoadTestSQLData(db, ctx)
	if err != nil {
		t.Errorf("failed to load SQL test data: %v", err)
	}
	err = loadSQLData(db, ctx, "./tests/does-not-exist.sql")
	if err == nil {
		t.Errorf("did not fail to load SQL test data")
	}
	err = loadTestSQLDataFiles(db, ctx, []string{"./tests/does-not-exist.sql"})
	if err == nil {
		t.Errorf("did not fail to load SQL test data")
	}
	err = loadSQLData(db, ctx, "./tests/bad_sql.sql")
	if err == nil {
		t.Errorf("did not fail to load SQL test data")
	}
}
