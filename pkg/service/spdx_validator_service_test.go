package service

import "testing"

func TestLicenseValidator(t *testing.T) {
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
			validator := GetSPDXValidator()
			if !validator.IsValidLicenseID(tt.licenseID) && !tt.expectErr {
				t.Fatalf("Expected license ID '%s' to be valid, but got false", tt.licenseID)
			}
		})
	}
}
