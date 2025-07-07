package middleware

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/scanoss/papi/api/licensesv2"
	"scanoss.com/licenses/pkg/dto"
	"testing"
)

func TestLicenseDetailsMiddleware(t *testing.T) {
	ctx := context.Background()
	s := ctxzap.Extract(ctx).Sugar()
	tests := []struct {
		name      string
		licenseID string
		expectErr bool
	}{
		{
			name:      "should process MIT license",
			licenseID: "MIT",
			expectErr: false,
		},
		{
			name:      "should process Apache license",
			licenseID: "Apache-2.0",
			expectErr: false,
		},
		{
			name:      "should process GPL license",
			licenseID: "GPL-3.0",
			expectErr: false,
		},
		{
			name:      "should handle empty license ID",
			licenseID: "",
			expectErr: true,
		},
		{
			name:      "should handle invalid license ID",
			licenseID: "INVALID-LICENSE",
			expectErr: true,
		},
		{
			name:      "should process license with custom SPDX-ID",
			licenseID: "LicenseRef-custom-spdx",
			expectErr: false,
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

			_, err := middleware.Process()

			if tt.expectErr {
				if err == nil {
					t.Fatalf("Expected error for license ID '%s', but got nil", tt.licenseID)
				}

			}
		})
	}
}
