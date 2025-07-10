package usecase

import (
	"context"
	"errors"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/licensesv2"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	"github.com/stretchr/testify/mock"
	myconfig "scanoss.com/licenses/pkg/config"
	"scanoss.com/licenses/pkg/dto"
	models "scanoss.com/licenses/pkg/model"
	"scanoss.com/licenses/pkg/protocol/rest"
	"testing"
)

// Mock implementation using testify/mock
type MockLicenseModel struct {
	mock.Mock
}

func (m *MockLicenseModel) GetLicenseByID(id string) (models.License, error) {
	args := m.Called(id)
	return args.Get(0).(models.License), args.Error(1)
}

func TestLicenseUseCase_GetDetails(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()
	config := &myconfig.ServerConfig{}
	defer zlog.SyncZap()

	tests := []struct {
		name           string
		licModel       models.LicenseModelInterface
		licenseRequest dto.LicenseRequestDTO
		expectedResult pb.LicenseResponse
		expectedError  *Error
		expectErr      bool
	}{
		{
			name: "successful license retrieval",
			licModel: func() models.LicenseModelInterface {
				mockModel := new(MockLicenseModel)
				mockModel.On("GetLicenseByID", "MIT").Return(models.License{
					ID:                    1,
					Name:                  "MIT License",
					LicenseId:             "MIT",
					DetailsUrl:            "https://spdx.org/licenses/MIT.html",
					Reference:             "https://opensource.org/licenses/MIT",
					IsDeprecatedLicenseId: false,
					IsOsiApproved:         true,
					SeeAlso:               models.SeeAlso{"https://opensource.org/licenses/MIT"},
					IsFsfLibre:            true,
				}, nil)
				return mockModel
			}(),
			licenseRequest: dto.LicenseRequestDTO{
				ID: "MIT",
			},
			expectedResult: pb.LicenseResponse{
				FullName: "MIT License",
				Spdx: &pb.SPDX{
					FullName:      "MIT License",
					Id:            "MIT",
					DetailsUrl:    "https://spdx.org/licenses/MIT.html",
					ReferenceUrl:  "https://opensource.org/licenses/MIT",
					IsDeprecated:  false,
					IsOsiApproved: true,
					SeeAlso:       []string{"https://opensource.org/licenses/MIT"},
					IsFsfLibre:    true,
				},
			},
			expectedError: nil,
			expectErr:     false,
		},
		{
			name: "license not found",
			licModel: func() models.LicenseModelInterface {
				mockModel := new(MockLicenseModel)
				mockModel.On("GetLicenseByID", "NONEXISTENT").Return(models.License{}, nil)
				return mockModel
			}(),
			licenseRequest: dto.LicenseRequestDTO{
				ID: "NONEXISTENT",
			},
			expectedResult: pb.LicenseResponse{},
			expectedError: &Error{
				Status:  common.StatusCode_SUCCEEDED_WITH_WARNINGS,
				Code:    rest.HTTP_CODE_404,
				Message: "License not found",
				Error:   errors.New("license not found"),
			},
			expectErr: true,
		},
		{
			name: "license found but with ID 0 (empty result)",
			licModel: func() models.LicenseModelInterface {
				mockModel := new(MockLicenseModel)
				mockModel.On("GetLicenseByID", "EMPTY").Return(models.License{}, nil)
				return mockModel
			}(),
			licenseRequest: dto.LicenseRequestDTO{
				ID: "EMPTY",
			},
			expectedResult: pb.LicenseResponse{},
			expectedError: &Error{
				Status:  common.StatusCode_SUCCEEDED_WITH_WARNINGS,
				Code:    rest.HTTP_CODE_404,
				Message: "License not found",
				Error:   errors.New("license not found"),
			},
			expectErr: true,
		},
		{
			name: "case insensitive license ID",
			licModel: func() models.LicenseModelInterface {
				mockModel := new(MockLicenseModel)
				mockModel.On("GetLicenseByID", "mit").Return(models.License{
					ID:                    1,
					Name:                  "MIT License",
					LicenseId:             "MIT",
					DetailsUrl:            "https://spdx.org/licenses/MIT.html",
					Reference:             "https://opensource.org/licenses/MIT",
					IsDeprecatedLicenseId: false,
					IsOsiApproved:         true,
					SeeAlso:               models.SeeAlso{"https://opensource.org/licenses/MIT"},
					IsFsfLibre:            true,
				}, nil)
				return mockModel
			}(),
			licenseRequest: dto.LicenseRequestDTO{
				ID: "mit",
			},
			expectedResult: pb.LicenseResponse{
				FullName: "MIT License",
				Spdx: &pb.SPDX{
					FullName:      "MIT License",
					Id:            "MIT",
					DetailsUrl:    "https://spdx.org/licenses/MIT.html",
					ReferenceUrl:  "https://opensource.org/licenses/MIT",
					IsDeprecated:  false,
					IsOsiApproved: true,
					SeeAlso:       []string{"https://opensource.org/licenses/MIT"},
					IsFsfLibre:    true,
				},
			},
			expectedError: nil,
			expectErr:     false,
		},
		{
			name: "error db model",
			licModel: func() models.LicenseModelInterface {
				mockModel := new(MockLicenseModel)
				mockModel.On("GetLicenseByID", "MIT").Return(models.License{}, errors.New("error connecting to db"))
				return mockModel
			}(),
			licenseRequest: dto.LicenseRequestDTO{
				ID: "MIT",
			},
			expectedResult: pb.LicenseResponse{}, // Should be empty on error
			expectedError: &Error{
				Status:  common.StatusCode_FAILED,
				Code:    rest.HTTP_CODE_500,
				Message: "error connecting to db",
				Error:   errors.New("error connecting to db"),
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := NewLicenseUseCaseWithLicenseModel(ctx, s, config, tt.licModel)
			_, usecaseErr := usecase.GetDetails(tt.licenseRequest)
			if tt.expectErr {
				if usecaseErr == nil {
					{
						t.Fatalf("Expected error, but got %v", tt.expectedError)
					}
				}
			}
		})
	}
}
