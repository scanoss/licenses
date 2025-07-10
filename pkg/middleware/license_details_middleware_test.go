package middleware

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/scanoss/papi/api/licensesv2"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	"scanoss.com/licenses/pkg/dto"
	"testing"
)

func TestLicenseDetailsMiddleware(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()
	tests := []struct {
		name      string
		licenseID string
		expected  dto.LicenseRequestDTO
		expectErr bool
	}{
		{
			name:      "should process MIT license",
			licenseID: "MIT",
			expected:  dto.LicenseRequestDTO{ID: "MIT"},
			expectErr: false,
		},
		{
			name:      "should process Apache license",
			licenseID: "Apache-2.0",
			expected:  dto.LicenseRequestDTO{ID: "Apache-2.0"},
			expectErr: false,
		},
		{
			name:      "should process GPL license",
			licenseID: "GPL-3.0",
			expected:  dto.LicenseRequestDTO{ID: "GPL-3.0"},
			expectErr: false,
		},
		{
			name:      "should not handle empty license ID",
			licenseID: "",
			expected:  dto.LicenseRequestDTO{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := &LicenseDetailMiddleware[[]dto.LicenseRequestDTO]{
				MiddlewareBase: MiddlewareBase{s: s},
				req: &licensesv2.LicenseRequest{
					Id: tt.licenseID,
				},
			}

			licenseDTO, err := middleware.Process()

			if tt.expectErr {
				if err == nil {
					t.Fatalf("Expected error for license ID '%s', but got nil", tt.licenseID)
				}
			}

			if licenseDTO != tt.expected {
				t.Fatalf("Expected license ID '%s', but got '%s'", tt.licenseID, licenseDTO.ID)
			}

		})
	}
}
