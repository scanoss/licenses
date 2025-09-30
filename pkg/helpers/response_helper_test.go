package helpers

import (
	"net/http"
	"testing"

	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/licensesv2"
)

func TestComponentsLicenseInfoStatus(t *testing.T) {
	tests := []struct {
		name                string
		componentLicenses   []*pb.ComponentLicenseInfo
		expectedGRPCStatus  common.StatusCode
		expectedHTTPCode    int
		expectedMessagePart string
	}{
		{
			name:                "nil component licenses",
			componentLicenses:   nil,
			expectedGRPCStatus:  common.StatusCode_FAILED,
			expectedHTTPCode:    http.StatusNotFound,
			expectedMessagePart: "Licenses not found for requested component(s)",
		},
		{
			name:                "all components without licenses",
			componentLicenses:   []*pb.ComponentLicenseInfo{{Purl: "pkg:npm/test@1.0.0", Licenses: []*pb.LicenseInfo{}}},
			expectedGRPCStatus:  common.StatusCode_FAILED,
			expectedHTTPCode:    http.StatusNotFound,
			expectedMessagePart: "No licenses found for the following component(s):",
		},
		{
			name: "all components with licenses",
			componentLicenses: []*pb.ComponentLicenseInfo{
				{Purl: "pkg:npm/test@1.0.0", Licenses: []*pb.LicenseInfo{{Id: "MIT"}}},
			},
			expectedGRPCStatus:  common.StatusCode_SUCCESS,
			expectedHTTPCode:    http.StatusOK,
			expectedMessagePart: "Licenses retrieved successfully",
		},
		{
			name: "mixed results with warnings",
			componentLicenses: []*pb.ComponentLicenseInfo{
				{Purl: "pkg:npm/test@1.0.0", Licenses: []*pb.LicenseInfo{{Id: "MIT"}}},
				{Purl: "pkg:npm/nolicense@1.0.0", Licenses: []*pb.LicenseInfo{}},
			},
			expectedGRPCStatus:  common.StatusCode_SUCCEEDED_WITH_WARNINGS,
			expectedHTTPCode:    http.StatusOK,
			expectedMessagePart: "No licenses found for the following component(s):",
		},
		{
			name: "multiple components with licenses",
			componentLicenses: []*pb.ComponentLicenseInfo{
				{Purl: "pkg:npm/test1@1.0.0", Licenses: []*pb.LicenseInfo{{Id: "MIT"}}},
				{Purl: "pkg:npm/test2@1.0.0", Licenses: []*pb.LicenseInfo{{Id: "Apache-2.0"}}},
			},
			expectedGRPCStatus:  common.StatusCode_SUCCESS,
			expectedHTTPCode:    http.StatusOK,
			expectedMessagePart: "Licenses retrieved successfully",
		},
		{
			name: "more not found than success",
			componentLicenses: []*pb.ComponentLicenseInfo{
				{Purl: "pkg:npm/test@1.0.0", Licenses: []*pb.LicenseInfo{{Id: "MIT"}}},
				{Purl: "pkg:npm/nolicense1@1.0.0", Licenses: []*pb.LicenseInfo{}},
				{Purl: "pkg:npm/nolicense2@1.0.0", Licenses: []*pb.LicenseInfo{}},
			},
			expectedGRPCStatus:  common.StatusCode_SUCCEEDED_WITH_WARNINGS,
			expectedHTTPCode:    http.StatusOK,
			expectedMessagePart: "Licenses retrieved successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gRPCStatus, httpCode, message := componentsLicenseInfoStatus(tt.componentLicenses)

			if gRPCStatus != tt.expectedGRPCStatus {
				t.Errorf("expected gRPC status %v, got %v", tt.expectedGRPCStatus, gRPCStatus)
			}
			if httpCode != tt.expectedHTTPCode {
				t.Errorf("expected HTTP code %d, got %d", tt.expectedHTTPCode, httpCode)
			}
			if tt.expectedMessagePart != "" && len(message) == 0 {
				t.Errorf("expected message to contain '%s', got empty message", tt.expectedMessagePart)
			}
		})
	}
}

func TestComponentLicenseInfoStatus(t *testing.T) {
	tests := []struct {
		name               string
		componentLicense   *pb.ComponentLicenseInfo
		expectedGRPCStatus common.StatusCode
		expectedHTTPCode   int
	}{
		{
			name:               "nil component license",
			componentLicense:   nil,
			expectedGRPCStatus: common.StatusCode_FAILED,
			expectedHTTPCode:   http.StatusNotFound,
		},
		{
			name:               "component without licenses",
			componentLicense:   &pb.ComponentLicenseInfo{Purl: "pkg:npm/test@1.0.0", Licenses: []*pb.LicenseInfo{}},
			expectedGRPCStatus: common.StatusCode_FAILED,
			expectedHTTPCode:   http.StatusNotFound,
		},
		{
			name: "component with licenses",
			componentLicense: &pb.ComponentLicenseInfo{
				Purl:     "pkg:npm/test@1.0.0",
				Licenses: []*pb.LicenseInfo{{Id: "MIT"}},
			},
			expectedGRPCStatus: common.StatusCode_SUCCESS,
			expectedHTTPCode:   http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gRPCStatus, httpCode, _ := componentLicenseInfoStatus(tt.componentLicense)

			if gRPCStatus != tt.expectedGRPCStatus {
				t.Errorf("expected gRPC status %v, got %v", tt.expectedGRPCStatus, gRPCStatus)
			}
			if httpCode != tt.expectedHTTPCode {
				t.Errorf("expected HTTP code %d, got %d", tt.expectedHTTPCode, httpCode)
			}
		})
	}
}

func TestLicenseDetailsStatus(t *testing.T) {
	tests := []struct {
		name               string
		licenseDetails     *pb.LicenseDetails
		expectedGRPCStatus common.StatusCode
		expectedHTTPCode   int
		expectedMessage    string
	}{
		{
			name:               "nil license details",
			licenseDetails:     nil,
			expectedGRPCStatus: common.StatusCode_FAILED,
			expectedHTTPCode:   http.StatusNotFound,
			expectedMessage:    "License details not found",
		},
		{
			name:               "valid license details",
			licenseDetails:     &pb.LicenseDetails{FullName: "MIT License"},
			expectedGRPCStatus: common.StatusCode_SUCCESS,
			expectedHTTPCode:   http.StatusOK,
			expectedMessage:    "License details retrieved successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gRPCStatus, httpCode, message := licenseDetailsStatus(tt.licenseDetails)

			if gRPCStatus != tt.expectedGRPCStatus {
				t.Errorf("expected gRPC status %v, got %v", tt.expectedGRPCStatus, gRPCStatus)
			}
			if httpCode != tt.expectedHTTPCode {
				t.Errorf("expected HTTP code %d, got %d", tt.expectedHTTPCode, httpCode)
			}
			if message != tt.expectedMessage {
				t.Errorf("expected message '%s', got '%s'", tt.expectedMessage, message)
			}
		})
	}
}

func TestDetermineStatusResponse(t *testing.T) {
	tests := []struct {
		name               string
		data               interface{}
		expectedGRPCStatus common.StatusCode
		expectedHTTPCode   int
		expectedMessage    string
	}{
		{
			name: "slice of ComponentLicenseInfo",
			data: []*pb.ComponentLicenseInfo{
				{Purl: "pkg:npm/test@1.0.0", Licenses: []*pb.LicenseInfo{{Id: "MIT"}}},
			},
			expectedGRPCStatus: common.StatusCode_SUCCESS,
			expectedHTTPCode:   http.StatusOK,
			expectedMessage:    "Licenses retrieved successfully",
		},
		{
			name: "single ComponentLicenseInfo",
			data: &pb.ComponentLicenseInfo{
				Purl:     "pkg:npm/test@1.0.0",
				Licenses: []*pb.LicenseInfo{{Id: "MIT"}},
			},
			expectedGRPCStatus: common.StatusCode_SUCCESS,
			expectedHTTPCode:   http.StatusOK,
			expectedMessage:    "Licenses retrieved successfully",
		},
		{
			name:               "LicenseDetails",
			data:               &pb.LicenseDetails{FullName: "MIT License"},
			expectedGRPCStatus: common.StatusCode_SUCCESS,
			expectedHTTPCode:   http.StatusOK,
			expectedMessage:    "License details retrieved successfully",
		},
		{
			name:               "unsupported type",
			data:               "invalid type",
			expectedGRPCStatus: common.StatusCode_FAILED,
			expectedHTTPCode:   http.StatusInternalServerError,
			expectedMessage:    "Internal server error",
		},
		{
			name:               "nil value",
			data:               nil,
			expectedGRPCStatus: common.StatusCode_FAILED,
			expectedHTTPCode:   http.StatusInternalServerError,
			expectedMessage:    "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gRPCStatus, httpCode, message := DetermineStatusResponse(tt.data)

			if gRPCStatus != tt.expectedGRPCStatus {
				t.Errorf("expected gRPC status %v, got %v", tt.expectedGRPCStatus, gRPCStatus)
			}
			if httpCode != tt.expectedHTTPCode {
				t.Errorf("expected HTTP code %d, got %d", tt.expectedHTTPCode, httpCode)
			}
			if message != tt.expectedMessage {
				t.Errorf("expected message '%s', got '%s'", tt.expectedMessage, message)
			}
		})
	}
}
