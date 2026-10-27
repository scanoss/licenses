package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/licensesv2"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"net/http"
	myconfig "scanoss.com/licenses/pkg/config"
	"scanoss.com/licenses/pkg/dto"
	models "scanoss.com/licenses/pkg/model"
	"testing"

	_ "modernc.org/sqlite"
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

func (m *MockOSADLModel) GetOSADLByLicenseID(ctx context.Context, s *zap.SugaredLogger, id string) (models.OSADL, error) {
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
					Name:                  "MIT LicenseDetail",
					LicenseID:             "MIT",
					DetailsURL:            "https://spdx.org/licenses/MIT.html",
					Reference:             "https://opensource.org/licenses/MIT",
					IsDeprecatedLicenseID: false,
					IsOsiApproved:         true,
					SeeAlso:               pq.StringArray{"https://opensource.org/licenses/MIT"},
					IsFsfLibre:            true,
				}, nil)
				return mockModel
			}(),
			osadlModel: func() models.OSADLModelInterface {
				mockModel := new(MockOSADLModel)
				mockModel.On("GetOSADLByLicenseID", "MIT").Return(models.OSADL{
					ID:                       "MIT",
					Compatibilities:          pq.StringArray{},
					Incompatibilities:        pq.StringArray{},
					DependingCompatibilities: pq.StringArray{},
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
				mockModel.On("GetOSADLByLicenseID", "NONEXISTENT").Return(models.OSADL{}, nil)
				return mockModel
			}(),
			licenseRequest: dto.LicenseRequestDTO{
				ID: "NONEXISTENT",
			},
			expectedResult: pb.LicenseDetails{},
			expectedError: &Error{
				Status:  common.StatusCode_SUCCEEDED_WITH_WARNINGS,
				Code:    http.StatusNotFound,
				Message: "LicenseDetail not found",
				Error:   errors.New("license not found"),
			},
			expectErr: true,
		},
		{
			name: "license found but with empty LicenseID (empty result)",
			licModel: func() models.LicenseDetailModelInterface {
				mockModel := new(MockLicenseModel)
				mockModel.On("GetLicenseByID", "EMPTY").Return(models.LicenseDetail{}, nil)
				return mockModel
			}(),
			osadlModel: func() models.OSADLModelInterface {
				mockModel := new(MockOSADLModel)
				mockModel.On("GetOSADLByLicenseID", "EMPTY").Return(models.OSADL{}, nil)
				return mockModel
			}(),
			licenseRequest: dto.LicenseRequestDTO{
				ID: "EMPTY",
			},
			expectedResult: pb.LicenseDetails{},
			expectedError: &Error{
				Status:  common.StatusCode_SUCCEEDED_WITH_WARNINGS,
				Code:    http.StatusNotFound,
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
					Name:                  "MIT LicenseDetail",
					LicenseID:             "MIT",
					DetailsURL:            "https://spdx.org/licenses/MIT.html",
					Reference:             "https://opensource.org/licenses/MIT",
					IsDeprecatedLicenseID: false,
					IsOsiApproved:         true,
					SeeAlso:               pq.StringArray{"https://opensource.org/licenses/MIT"},
					IsFsfLibre:            true,
				}, nil)
				return mockModel
			}(),
			osadlModel: func() models.OSADLModelInterface {
				mockModel := new(MockOSADLModel)
				mockModel.On("GetOSADLByLicenseID", "MIT").Return(models.OSADL{
					ID:                       "MIT",
					Compatibilities:          pq.StringArray{},
					Incompatibilities:        pq.StringArray{},
					DependingCompatibilities: pq.StringArray{},
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
				mockModel.On("GetOSADLByLicenseID", "NONEXISTENT").Return(models.OSADL{}, errors.New("error connecting to db"))
				return mockModel
			}(),
			licenseRequest: dto.LicenseRequestDTO{
				ID: "MIT",
			},
			expectedResult: pb.LicenseDetails{}, // Should be empty on error
			expectedError: &Error{
				Status:  common.StatusCode_FAILED,
				Code:    http.StatusInternalServerError,
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
					Name:                  "MIT LicenseDetail",
					LicenseID:             "MIT",
					DetailsURL:            "https://spdx.org/licenses/MIT.html",
					Reference:             "https://opensource.org/licenses/MIT",
					IsDeprecatedLicenseID: false,
					IsOsiApproved:         true,
					SeeAlso:               pq.StringArray{"https://opensource.org/licenses/MIT"},
					IsFsfLibre:            true,
				}, nil)
				return mockModel
			}(),
			osadlModel: func() models.OSADLModelInterface {
				mockModel := new(MockOSADLModel)
				mockModel.On("GetOSADLByLicenseID", "MIT").Return(models.OSADL{}, errors.New("error connecting to osadl db"))
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

// The fixture's pkg:cargo/serde has a licence in component_license and none in
// component_version_license; pkg:npm/express has both, and its per-release licences
// must win.
const (
	serdePurlID   = "3d9c7a14-2b6e-5f08-9c41-7ae0d5b12f63"
	expressPurlID = "eea6a5f0-ac07-5faa-86d6-42a1f88f80bd"
)

func setupPackageFallbackUseCase(t *testing.T, sourcePriority []int16) (context.Context, LicenseUseCase, func()) {
	t.Helper()
	if err := zlog.NewSugaredDevLogger(); err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	db, err := sqlx.Connect("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("error connecting to DB: %v", err)
	}
	if err = models.LoadTestSQLData(db, ctx); err != nil {
		t.Fatalf("error loading test SQL data: %v", err)
	}
	config := &myconfig.ServerConfig{}
	config.Lookup.SourcePriority = sourcePriority
	return ctx, *NewLicenseUseCase(config, db, nil), func() {
		models.CloseDB(db)
		zlog.SyncZap()
	}
}

func TestFetchPackageLicenceFallback(t *testing.T) {
	declaredFirst := []int16{0, 31, 32}

	t.Run("a component with no per-version licence falls back", func(t *testing.T) {
		ctx, lu, cleanup := setupPackageFallbackUseCase(t, declaredFirst)
		defer cleanup()
		s := ctxzap.Extract(ctx).Sugar()

		got := lu.fetchLicensesByPurlAndVersion(ctx, s, serdePurlID, "1.0.0")
		if len(got) != 2 {
			t.Fatalf("got %d licences, want 2 from component_license", len(got))
		}
		for _, l := range got {
			// The package-level licence belongs to no release, so the row must not
			// claim the version the caller asked about.
			if l.Version != "" {
				t.Errorf("version = %q, want empty", l.Version)
			}
		}
	})

	t.Run("per-version licences are not overridden", func(t *testing.T) {
		ctx, lu, cleanup := setupPackageFallbackUseCase(t, []int16{1, 2, 0})
		defer cleanup()
		s := ctxzap.Extract(ctx).Sugar()

		got := lu.fetchLicensesByPurlAndVersion(ctx, s, expressPurlID, "4.18.2")
		if len(got) == 0 {
			t.Fatal("no licences returned for a component that has per-version rows")
		}
		for _, l := range got {
			if l.Version != "4.18.2" {
				t.Errorf("version = %q, want 4.18.2: the fallback displaced real data", l.Version)
			}
			// 15 is express's component_license row. Seeing it here means the
			// fallback ran when it should not have.
			if l.LicenseID == 15 {
				t.Error("package-level licence 15 returned although the release has its own")
			}
		}
	})

	t.Run("a priority list without the declared source does not fall back", func(t *testing.T) {
		ctx, lu, cleanup := setupPackageFallbackUseCase(t, []int16{31, 32})
		defer cleanup()
		s := ctxzap.Extract(ctx).Sugar()

		if got := lu.fetchLicensesByPurlAndVersion(ctx, s, serdePurlID, "1.0.0"); len(got) != 0 {
			t.Errorf("got %d licences, want none: source 0 was not requested", len(got))
		}
	})

	t.Run("an unversioned lookup falls back too", func(t *testing.T) {
		ctx, lu, cleanup := setupPackageFallbackUseCase(t, declaredFirst)
		defer cleanup()
		s := ctxzap.Extract(ctx).Sugar()

		if got := lu.fetchLicensesByPurl(ctx, s, serdePurlID, declaredFirst); len(got) != 2 {
			t.Errorf("got %d licences, want 2 from component_license", len(got))
		}
	})
}
