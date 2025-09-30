package models

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	"testing"
)

func TestGetOSDALByLicenseId(t *testing.T) {
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
	err = loadTestSQLDataFiles(db, ctx, []string{"tests/osadl.sql"})
	if err != nil {
		t.Fatalf("failed to load SQL test data: %v", err)
	}
	osadlModel := NewOSADLModel(db)

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
			license, err := osadlModel.GetOSADLByLicenseId(ctx, s, test.licenseID)
			if test.expectErr {
				if err == nil {
					t.Errorf("osadlModel.GetOSADLByLicenseId() error = %v, wantErr %v", err, test.expectErr)
				}
			}
			fmt.Printf("LicenseDetail: %#v\n", license)
		})
	}
}
