package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-component-helper/componenthelper"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	"net/http"
	"os"
	models "scanoss.com/licenses/pkg/model"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	common "github.com/scanoss/papi/api/commonv2"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	myconfig "scanoss.com/licenses/pkg/config"
	"scanoss.com/licenses/pkg/dto"
)

type mockMiddleware struct {
	processFunc func() ([]componenthelper.ComponentDTO, error)
}

func (m *mockMiddleware) Process() ([]componenthelper.ComponentDTO, error) {
	return m.processFunc()
}

type mockComponentMiddleware struct {
	processFunc func() (componenthelper.ComponentDTO, error)
}

func (m *mockComponentMiddleware) Process() (componenthelper.ComponentDTO, error) {
	return m.processFunc()
}

type mockLicenseDetailsMiddleware struct {
	processFunc func() (dto.LicenseRequestDTO, error)
}

func (m *mockLicenseDetailsMiddleware) Process() (dto.LicenseRequestDTO, error) {
	return m.processFunc()
}

func TestNewLicenseHandler(t *testing.T) {
	config := &myconfig.ServerConfig{}
	handler := NewLicenseHandler(config, &sqlx.DB{}, nil)

	if handler == nil {
		t.Fatal("Expected handler to be created, got nil")
	}

	if handler.config != config {
		t.Error("Expected handler config to be set correctly")
	}
}

func TestLicenseHandler_getResponseStatus(t *testing.T) {
	config := &myconfig.ServerConfig{}
	handler := NewLicenseHandler(config, &sqlx.DB{}, nil)
	ctx := context.Background()
	logger := zap.NewNop().Sugar()

	t.Run("success case", func(t *testing.T) {
		ctx = metadata.NewIncomingContext(ctx, metadata.New(map[string]string{}))
		status := handler.getResponseStatus(logger, ctx, common.StatusCode_SUCCESS, http.StatusOK, "Licenses Successfully retrieved", nil)

		if status.Status != common.StatusCode_SUCCESS {
			t.Errorf("Expected status SUCCESS, got %v", status.Status)
		}

		if status.Message != "Licenses Successfully retrieved" {
			t.Errorf("Expected success message, got %s", status.Message)
		}
	})

	t.Run("error case", func(t *testing.T) {
		ctx = metadata.NewIncomingContext(ctx, metadata.New(map[string]string{}))
		err := errors.New("test error")
		status := handler.getResponseStatus(logger, ctx, common.StatusCode_FAILED, http.StatusBadRequest, "", err)

		if status.Status != common.StatusCode_FAILED {
			t.Errorf("Expected status FAILED, got %v", status.Status)
		}

		if status.Message != "test error" {
			t.Errorf("Expected error message, got %s", status.Message)
		}
	})
}

func TestLicenseHandler_GetLicenses(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	config := &myconfig.ServerConfig{}
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	db, err := sqlx.Connect("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(fmt.Sprintf("Error connecting to DB %v", err))
	}
	err = models.LoadTestSQLData(db, ctx)
	if err != nil {
		t.Fatal(fmt.Sprintf("Error loading test SQL data %v", err))
	}
	defer models.CloseDB(db)
	handler := NewLicenseHandler(config, db, nil)
	t.Run("successful middleware processing", func(t *testing.T) {
		mockMW := &mockMiddleware{
			processFunc: func() ([]componenthelper.ComponentDTO, error) {
				return []componenthelper.ComponentDTO{
					{Purl: "pkg:gitlab/gpl/project", Requirement: "1.0.0"},
				}, nil
			},
		}

		response, err := handler.GetComponentsLicense(ctx, mockMW)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if response == nil {
			t.Fatal("Expected response, got nil")
		}

		if response.Status.Status != common.StatusCode_SUCCESS {
			t.Errorf("Expected SUCCESS status, got %v", response.Status.Status)
		}

	})
}

func TestLicenseHandler_GetComponentLicense(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	config := &myconfig.ServerConfig{}
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	db, err := sqlx.Connect("sqlite", "file::memory:?cache=shared")
	err = models.LoadTestSQLData(db, ctx)
	if err != nil {
		t.Fatal(fmt.Sprintf("Error loading test SQL data %v", err))
	}
	defer models.CloseDB(db)

	handler := NewLicenseHandler(config, db, nil)

	t.Run("successful middleware processing", func(t *testing.T) {
		mockMW := &mockComponentMiddleware{
			processFunc: func() (componenthelper.ComponentDTO, error) {
				return componenthelper.ComponentDTO{
					Purl:        "pkg:gitlab/gpl/project",
					Requirement: "1.0.0",
				}, nil
			},
		}

		response, err := handler.GetComponentLicense(ctx, mockMW)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if response == nil {
			t.Fatal("Expected response, got nil")
		}

		if response.Status.Status != common.StatusCode_SUCCESS {
			t.Errorf("Expected SUCCESS status, got %v", response.Status.Status)
		}

		if response.Component == nil {
			t.Error("Expected component to be initialized")
		}
	})

	t.Run("middleware processing error", func(t *testing.T) {
		mockMW := &mockComponentMiddleware{
			processFunc: func() (componenthelper.ComponentDTO, error) {
				return componenthelper.ComponentDTO{}, errors.New("middleware error")
			},
		}

		response, err := handler.GetComponentLicense(ctx, mockMW)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if response == nil {
			t.Fatal("Expected response, got nil")
		}

		if response.Status.Status != common.StatusCode_FAILED {
			t.Errorf("Expected FAILED status, got %v", response.Status.Status)
		}

		if response.Component == nil {
			t.Error("Expected component to be initialized")
		}
	})

}

func TestLicenseHandler_GetComponentsLicense_ResponseStatus(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	config := &myconfig.ServerConfig{}
	config.Lookup.SourcePriority = []int16{31, 32, 33, 5}
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	db, err := sqlx.Connect("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(fmt.Sprintf("Error connecting to DB %v", err))
	}
	err = models.LoadTestSQLData(db, ctx)
	if err != nil {
		t.Fatal(fmt.Sprintf("Error loading test SQL data %v", err))
	}
	defer models.CloseDB(db)
	handler := NewLicenseHandler(config, db, nil)

	tests := []struct {
		name             string
		component        componenthelper.ComponentDTO
		expectedStatus   common.StatusCode
		expectErrMessage bool
		expectErrCode    bool
	}{
		{
			name:             "unknown component returns error_message and error_code",
			component:        componenthelper.ComponentDTO{Purl: "pkg:npm/unknown/nonexistent-package", Requirement: "1.0.0"},
			expectedStatus:   common.StatusCode_SUCCESS,
			expectErrMessage: true,
			expectErrCode:    true,
		},
		{
			name:             "invalid purl returns error_code",
			component:        componenthelper.ComponentDTO{Purl: "not-a-valid-purl", Requirement: ""},
			expectedStatus:   common.StatusCode_SUCCESS,
			expectErrMessage: true,
			expectErrCode:    true,
		},
		{
			name:             "valid component has no error fields",
			component:        componenthelper.ComponentDTO{Purl: "pkg:gitlab/gpl/project", Requirement: "1.0.0"},
			expectedStatus:   common.StatusCode_SUCCESS,
			expectErrMessage: false,
			expectErrCode:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMW := &mockMiddleware{
				processFunc: func() ([]componenthelper.ComponentDTO, error) {
					return []componenthelper.ComponentDTO{tt.component}, nil
				},
			}

			response, err := handler.GetComponentsLicense(ctx, mockMW)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			if response == nil {
				t.Fatal("Expected response, got nil")
			}
			if response.Status.Status != tt.expectedStatus {
				t.Errorf("Expected status %v, got %v", tt.expectedStatus, response.Status.Status)
			}
			if len(response.Components) != 1 {
				t.Fatalf("Expected 1 component, got %d", len(response.Components))
			}

			comp := response.Components[0]
			if comp.Purl != tt.component.Purl {
				t.Errorf("Expected purl %s, got %s", tt.component.Purl, comp.Purl)
			}
			if tt.expectErrMessage && comp.ErrorMessage == nil {
				t.Errorf("Component %s: expected error_message to be set", comp.Purl)
			}
			if !tt.expectErrMessage && comp.ErrorMessage != nil {
				t.Errorf("Component %s: expected no error_message, got %q", comp.Purl, *comp.ErrorMessage)
			}
			if tt.expectErrCode && comp.ErrorCode == nil {
				t.Errorf("Component %s: expected error_code to be set", comp.Purl)
			}
			if !tt.expectErrCode && comp.ErrorCode != nil {
				t.Errorf("Component %s: expected no error_code, got %v", comp.Purl, *comp.ErrorCode)
			}
		})
	}
}

func TestLicenseHandler_GetComponentLicense_ResponseStatus(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	config := &myconfig.ServerConfig{}
	// Set look up priority
	config.Lookup.SourcePriority = []int16{31, 32, 33, 5}
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	db, err := sqlx.Connect("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(fmt.Sprintf("Error connecting to DB %v", err))
	}
	err = models.LoadTestSQLData(db, ctx)
	if err != nil {
		t.Fatal(fmt.Sprintf("Error loading test SQL data %v", err))
	}
	defer models.CloseDB(db)
	handler := NewLicenseHandler(config, db, nil)

	tests := []struct {
		name             string
		component        componenthelper.ComponentDTO
		expectedStatus   common.StatusCode
		expectErrMessage bool
		expectErrCode    bool
	}{
		{
			name:             "unknown component returns error fields",
			component:        componenthelper.ComponentDTO{Purl: "pkg:npm/unknown/nonexistent-package", Requirement: "1.0.0"},
			expectedStatus:   common.StatusCode_SUCCESS,
			expectErrMessage: true,
			expectErrCode:    true,
		},
		{
			name:             "invalid purl returns error fields",
			component:        componenthelper.ComponentDTO{Purl: "not-a-valid-purl"},
			expectedStatus:   common.StatusCode_SUCCESS,
			expectErrMessage: true,
			expectErrCode:    true,
		},
		{
			name:             "valid component has no error fields",
			component:        componenthelper.ComponentDTO{Purl: "pkg:gitlab/gpl/project", Requirement: "1.0.0"},
			expectedStatus:   common.StatusCode_SUCCESS,
			expectErrMessage: false,
			expectErrCode:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMW := &mockComponentMiddleware{
				processFunc: func() (componenthelper.ComponentDTO, error) {
					return tt.component, nil
				},
			}

			response, err := handler.GetComponentLicense(ctx, mockMW)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			if response == nil {
				t.Fatal("Expected response, got nil")
			}
			if response.Status.Status != tt.expectedStatus {
				t.Errorf("Expected status %v, got %v", tt.expectedStatus, response.Status.Status)
			}
			if response.Component.Purl != tt.component.Purl {
				t.Errorf("Expected purl %s, got %s", tt.component.Purl, response.Component.Purl)
			}
			if tt.expectErrMessage && response.Component.ErrorMessage == nil {
				t.Error("Expected error_message to be set")
			}
			if !tt.expectErrMessage && response.Component.ErrorMessage != nil {
				t.Errorf("Expected no error_message, got %q", *response.Component.ErrorMessage)
			}
			if tt.expectErrCode && response.Component.ErrorCode == nil {
				t.Error("Expected error_code to be set")
			}
			if !tt.expectErrCode && response.Component.ErrorCode != nil {
				t.Errorf("Expected no error_code, got %v", *response.Component.ErrorCode)
			}
		})
	}
}

func TestLicenseHandler_GetDetails(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	config := &myconfig.ServerConfig{}
	content, err := os.ReadFile("../model/tests/licenses.sql")
	if err != nil {
		t.Fatalf("Error reading SQL file: %v", err)
	}
	db, err := sqlx.Connect("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer models.CloseDB(db)
	_, err = db.Exec(string(content))
	if err != nil {
		t.Fatalf("Error reading SQL file: %v", err)
	}
	handler := NewLicenseHandler(config, db, nil)
	ctx := ctxzap.ToContext(context.Background(), zap.NewNop())

	t.Run("middleware processing error", func(t *testing.T) {
		mockMW := &mockLicenseDetailsMiddleware{
			processFunc: func() (dto.LicenseRequestDTO, error) {
				return dto.LicenseRequestDTO{}, errors.New("middleware error")
			},
		}

		response, err := handler.GetDetails(ctx, mockMW)

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if response == nil {
			t.Fatal("Expected response, got nil")
		}

		if response.Status.Status != common.StatusCode_FAILED {
			t.Errorf("Expected FAILED status, got %v", response.Status.Status)
		}

		if response.License == nil {
			t.Error("Expected license to be initialized")
		}
	})
}
