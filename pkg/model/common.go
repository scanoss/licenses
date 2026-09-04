// SPDX-LicenseDetail-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2018-2023 SCANOSS.COM
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

// This file common tasks for the models package

package models

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	_ "modernc.org/sqlite"
)

// loadSQLData Load the specified SQL files into the supplied DB.
func loadSQLData(db *sqlx.DB, ctx context.Context, filename string) error {
	fmt.Printf("Loading test data file: %v\n", filename)
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	if db != nil {
		_, err = db.ExecContext(ctx, string(file))
	}
	if err != nil {
		return err
	}
	return nil
}

// testDataFiles are the data fixtures loaded on top of the schema. They carry data
// only; the tables come from testSchemaDDL in test_schema.go.
var testDataFiles = []string{
	"../model/tests/component.sql",
	"../model/tests/component_version.sql",
	"../model/tests/component_version_license.sql",
	"../model/tests/license.sql",
	"../model/tests/spdx_license_data.sql",
	"../model/tests/osadl.sql",
	"../model/tests/osadl_lists.sql",
}

// idempotentDDL rewrites CREATE TABLE/INDEX into their IF NOT EXISTS form so a test
// can load the schema more than once. testSchemaDDL transcribes production and must
// not carry test-only clauses, so the rewrite happens here instead of in the constant.
func idempotentDDL(sql string) string {
	for _, kind := range []string{"TABLE", "INDEX"} {
		bare := "CREATE " + kind + " "
		guarded := bare + "IF NOT EXISTS "
		// Normalise first, so an already-guarded statement is not double-guarded.
		sql = strings.ReplaceAll(sql, guarded, bare)
		sql = strings.ReplaceAll(sql, bare, guarded)
	}
	return sql
}

// LoadTestSchema creates the serving schema in the supplied DB. Call this before
// loading any data fixture, since the fixtures do not define their own tables.
// It is safe to call more than once on the same DB.
func LoadTestSchema(db *sqlx.DB, ctx context.Context) error {
	if _, err := db.ExecContext(ctx, idempotentDDL(testSchemaDDL)); err != nil {
		return fmt.Errorf("failed to load the test schema: %w", err)
	}
	return nil
}

// LoadTestSQLData loads the serving schema plus all the test data fixtures.
func LoadTestSQLData(db *sqlx.DB, ctx context.Context) error {
	// In-memory SQLite gives each pooled connection its OWN empty database, and the
	// licence lookups run concurrently, so a second connection would find no tables.
	// Pin the pool to one so every query sees the same database.
	db.SetMaxOpenConns(1)
	if err := LoadTestSchema(db, ctx); err != nil {
		return err
	}
	return loadTestSQLDataFiles(db, ctx, testDataFiles)
}

// loadTestSQLDataFiles loads a list of test SQL files.
func loadTestSQLDataFiles(db *sqlx.DB, ctx context.Context, files []string) error {
	for _, file := range files {
		err := loadSQLData(db, ctx, file)
		if err != nil {
			return err
		}
	}
	return nil
}

// sqliteSetup sets up an in-memory SQL Lite DB for testing.
//
//nolint:unused // used by test files only
func sqliteSetup(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	return db
}

// sqliteConn sets up a connection to a test DB.
//
//nolint:unused // used by test files only
func sqliteConn(t *testing.T, ctx context.Context, db *sqlx.DB) *sqlx.Conn {
	conn, err := db.Connx(ctx) // Get a connection from the pool
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	return conn
}

// NewConn sets up a new database connection from the pool.
func NewConn(ctx context.Context, db *sqlx.DB) (*sqlx.Conn, error) {
	conn, err := db.Connx(ctx) // Get a connection from the pool
	if err != nil {
		errMsg := fmt.Sprintf("an error '%s' was not expected when opening a stub database connection", err)
		return &sqlx.Conn{}, errors.New(errMsg)
	}
	return conn, nil
}

// CloseDB closes the specified DB and logs any errors.
func CloseDB(db *sqlx.DB) {
	if db != nil {
		// zlog.S.Debugf("Closing DB...")
		err := db.Close()
		if err != nil {
			zlog.S.Warnf("Problem closing DB: %v", err)
		}
	}
}

// CloseConn closes the specified DB connection and logs any errors.
func CloseConn(conn *sqlx.Conn) {
	if conn != nil {
		// zlog.S.Debugf("Closing Connection...")
		err := conn.Close()
		if err != nil {
			zlog.S.Warnf("Problem closing DB connection: %v", err)
		}
	}
}

// CloseRows closes the specified DB query row and logs any errors.
func CloseRows(rows *sqlx.Rows) {
	if rows != nil {
		err := rows.Close()
		if err != nil {
			zlog.S.Warnf("Problem closing Rows: %v", err)
		}
	}
}
