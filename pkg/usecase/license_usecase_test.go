package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/licensesv2"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
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

type MockOSADLModel struct {
	mock.Mock
}

func (m *MockLicenseModel) GetLicenseByID(ctx context.Context, s *zap.SugaredLogger, id string) (models.LicenseDetail, error) {
	args := m.Called(id)
	return args.Get(0).(models.LicenseDetail), args.Error(1)
}

func (m *MockOSADLModel) GetOSADLByLicenseId(ctx context.Context, s *zap.SugaredLogger, id string) (models.OSADL, error) {
	args := m.Called(id)
	return args.Get(0).(models.OSADL), args.Error(1)
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
		licModel       models.LicenseDetailModelInterface
		osadlModel     models.OSADLModelInterface
		licenseRequest dto.LicenseRequestDTO
		expectedResult pb.LicenseDetails
		expectedError  *Error
		expectErr      bool
	}{
		{
			name: "successful license retrieval",
			licModel: func() models.LicenseDetailModelInterface {
				mockModel := new(MockLicenseModel)
				mockModel.On("GetLicenseByID", "MIT").Return(models.LicenseDetail{
					ID:                    1,
					Name:                  "MIT LicenseDetail",
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
			osadlModel: func() models.OSADLModelInterface {
				mockModel := new(MockOSADLModel)
				mockModel.On("GetOSADLByLicenseId", "MIT").Return(models.OSADL{
					ID:                       1,
					LicenseId:                "MIT",
					Compatibilities:          models.JSONStringSlice{},
					Incompatibilities:        models.JSONStringSlice{},
					DependingCompatibilities: models.JSONStringSlice{},
					PatentHints:              false,
					CopyleftClause:           false,
				}, nil)
				return mockModel
			}(),
			licenseRequest: dto.LicenseRequestDTO{
				ID: "MIT",
			},
			expectedResult: pb.LicenseDetails{
				FullName: "MIT LicenseDetail",
				Spdx: &pb.SPDX{
					FullName:      "MIT LicenseDetail",
					Id:            "MIT",
					DetailsUrl:    "https://spdx.org/licenses/MIT.html",
					ReferenceUrl:  "https://opensource.org/licenses/MIT",
					IsDeprecated:  false,
					IsOsiApproved: true,
					SeeAlso:       []string{"https://opensource.org/licenses/MIT"},
					IsFsfLibre:    true,
				},
				Osadl: &pb.OSADL{
					CopyleftClause:         false,
					Compatibility:          []string{""},
					Incompatibility:        []string{""},
					DependingCompatibility: []string{""},
					PatentHints:            false,
				},
			},
			expectedError: nil,
			expectErr:     false,
		},
		{
			name: "license not found",
			licModel: func() models.LicenseDetailModelInterface {
				mockModel := new(MockLicenseModel)
				mockModel.On("GetLicenseByID", "NONEXISTENT").Return(models.LicenseDetail{}, nil)
				return mockModel
			}(),
			osadlModel: func() models.OSADLModelInterface {
				mockModel := new(MockOSADLModel)
				mockModel.On("GetOSADLByLicenseId", "NONEXISTENT").Return(models.OSADL{}, nil)
				return mockModel
			}(),
			licenseRequest: dto.LicenseRequestDTO{
				ID: "NONEXISTENT",
			},
			expectedResult: pb.LicenseDetails{},
			expectedError: &Error{
				Status:  common.StatusCode_SUCCEEDED_WITH_WARNINGS,
				Code:    rest.HTTP_CODE_404,
				Message: "LicenseDetail not found",
				Error:   errors.New("license not found"),
			},
			expectErr: true,
		},
		{
			name: "license found but with ID 0 (empty result)",
			licModel: func() models.LicenseDetailModelInterface {
				mockModel := new(MockLicenseModel)
				mockModel.On("GetLicenseByID", "EMPTY").Return(models.LicenseDetail{}, nil)
				return mockModel
			}(),
			osadlModel: func() models.OSADLModelInterface {
				mockModel := new(MockOSADLModel)
				mockModel.On("GetOSADLByLicenseId", "EMPTY").Return(models.OSADL{}, nil)
				return mockModel
			}(),
			licenseRequest: dto.LicenseRequestDTO{
				ID: "EMPTY",
			},
			expectedResult: pb.LicenseDetails{},
			expectedError: &Error{
				Status:  common.StatusCode_SUCCEEDED_WITH_WARNINGS,
				Code:    rest.HTTP_CODE_404,
				Message: "LicenseDetail not found",
				Error:   errors.New("license not found"),
			},
			expectErr: true,
		},
		{
			name: "case insensitive license ID",
			licModel: func() models.LicenseDetailModelInterface {
				mockModel := new(MockLicenseModel)
				mockModel.On("GetLicenseByID", "mit").Return(models.LicenseDetail{
					ID:                    1,
					Name:                  "MIT LicenseDetail",
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
			osadlModel: func() models.OSADLModelInterface {
				mockModel := new(MockOSADLModel)
				mockModel.On("GetOSADLByLicenseId", "MIT").Return(models.OSADL{
					ID:                       1,
					LicenseId:                "MIT",
					Compatibilities:          models.JSONStringSlice{},
					Incompatibilities:        models.JSONStringSlice{},
					DependingCompatibilities: models.JSONStringSlice{},
					PatentHints:              false,
					CopyleftClause:           false,
				}, nil)
				return mockModel
			}(),
			licenseRequest: dto.LicenseRequestDTO{
				ID: "mit",
			},
			expectedResult: pb.LicenseDetails{
				FullName: "MIT LicenseDetail",
				Spdx: &pb.SPDX{
					FullName:      "MIT LicenseDetail",
					Id:            "MIT",
					DetailsUrl:    "https://spdx.org/licenses/MIT.html",
					ReferenceUrl:  "https://opensource.org/licenses/MIT",
					IsDeprecated:  false,
					IsOsiApproved: true,
					SeeAlso:       []string{"https://opensource.org/licenses/MIT"},
					IsFsfLibre:    true,
				},
				Osadl: &pb.OSADL{
					CopyleftClause:         false,
					Compatibility:          []string{""},
					Incompatibility:        []string{""},
					DependingCompatibility: []string{""},
					PatentHints:            false,
				},
			},
			expectedError: nil,
			expectErr:     false,
		},
		{
			name: "error license model",
			licModel: func() models.LicenseDetailModelInterface {
				mockModel := new(MockLicenseModel)
				mockModel.On("GetLicenseByID", "MIT").Return(models.LicenseDetail{}, errors.New("error connecting to db"))
				return mockModel
			}(),
			osadlModel: func() models.OSADLModelInterface {
				mockModel := new(MockOSADLModel)
				mockModel.On("GetOSADLByLicenseId", "NONEXISTENT").Return(models.OSADL{}, errors.New("error connecting to db"))
				return mockModel
			}(),
			licenseRequest: dto.LicenseRequestDTO{
				ID: "MIT",
			},
			expectedResult: pb.LicenseDetails{}, // Should be empty on error
			expectedError: &Error{
				Status:  common.StatusCode_FAILED,
				Code:    rest.HTTP_CODE_500,
				Message: "error connecting to db",
				Error:   errors.New("error connecting to db"),
			},
			expectErr: true,
		},
		{
			name: "error osadl model",
			licModel: func() models.LicenseDetailModelInterface {
				mockModel := new(MockLicenseModel)
				mockModel.On("GetLicenseByID", "MIT").Return(models.LicenseDetail{
					ID:                    1,
					Name:                  "MIT LicenseDetail",
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
			osadlModel: func() models.OSADLModelInterface {
				mockModel := new(MockOSADLModel)
				mockModel.On("GetOSADLByLicenseId", "MIT").Return(models.OSADL{}, errors.New("error connecting to osadl db"))
				return mockModel
			}(),
			licenseRequest: dto.LicenseRequestDTO{
				ID: "MIT",
			},
			expectedResult: pb.LicenseDetails{}, // Should be empty on error
			expectedError:  &Error{},
			expectErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := NewLicenseUseCaseWithLicenseModel(config, tt.licModel, tt.osadlModel)
			details, usecaseErr := usecase.GetDetails(ctx, s, tt.licenseRequest)
			if tt.expectErr {
				if usecaseErr == nil {
					{
						t.Fatalf("Expected error, but got %v", tt.expectedError)
					}
				}
			}
			fmt.Printf("Details: %#v\n", details)
		})
	}
}
