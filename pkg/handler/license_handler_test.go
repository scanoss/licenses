package handler

import (
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	"os"
	models "scanoss.com/licenses/pkg/model"
	"scanoss.com/licenses/pkg/protocol/rest"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	common "github.com/scanoss/papi/api/commonv2"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	myconfig "scanoss.com/licenses/pkg/config"
	"scanoss.com/licenses/pkg/dto"
)

type mockMiddleware struct {
	processFunc func() ([]dto.ComponentRequestDTO, error)
}

func (m *mockMiddleware) Process() ([]dto.ComponentRequestDTO, error) {
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
	handler := NewLicenseHandler(config, &sqlx.DB{})

	if handler == nil {
		t.Fatal("Expected handler to be created, got nil")
	}

	if handler.config != config {
		t.Error("Expected handler config to be set correctly")
	}
}

func TestLicenseHandler_getResponseStatus(t *testing.T) {
	config := &myconfig.ServerConfig{}
	handler := NewLicenseHandler(config, &sqlx.DB{})
	ctx := context.Background()
	logger := zap.NewNop().Sugar()

	t.Run("success case", func(t *testing.T) {
		ctx = metadata.NewIncomingContext(ctx, metadata.New(map[string]string{}))
		status := handler.getResponseStatus(logger, ctx, common.StatusCode_SUCCESS, rest.HTTP_CODE_200, nil)

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
		status := handler.getResponseStatus(logger, ctx, common.StatusCode_FAILED, rest.HTTP_CODE_400, err)

		if status.Status != common.StatusCode_FAILED {
			t.Errorf("Expected status FAILED, got %v", status.Status)
		}

		if status.Message != "test error" {
			t.Errorf("Expected error message, got %s", status.Message)
		}
	})
}

func TestLicenseHandler_GetLicenses(t *testing.T) {
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
	handler := NewLicenseHandler(config, db)
	ctx := context.Background()
	logger := ctxzap.Extract(ctx).Sugar()

	t.Run("successful middleware processing", func(t *testing.T) {
		mockMW := &mockMiddleware{
			processFunc: func() ([]dto.ComponentRequestDTO, error) {
				return []dto.ComponentRequestDTO{
					{Purl: "pkg:npm/test", Requirement: "1.0.0"},
				}, nil
			},
		}

		response, err := handler.GetLicenses(ctx, logger, mockMW)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if response == nil {
			t.Fatal("Expected response, got nil")
		}

		if response.Status.Status != common.StatusCode_SUCCESS {
			t.Errorf("Expected SUCCESS status, got %v", response.Status.Status)
		}

		if response.Licenses == nil {
			t.Error("Expected licenses array to be initialized")
		}
	})

	t.Run("middleware processing error", func(t *testing.T) {
		mockMW := &mockMiddleware{
			processFunc: func() ([]dto.ComponentRequestDTO, error) {
				return nil, errors.New("middleware error")
			},
		}

		response, err := handler.GetLicenses(ctx, logger, mockMW)

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if response == nil {
			t.Fatal("Expected response, got nil")
		}

		if response.Status.Status != common.StatusCode_FAILED {
			t.Errorf("Expected FAILED status, got %v", response.Status.Status)
		}

		if len(response.Licenses) != 0 {
			t.Errorf("Expected empty licenses array, got %d items", len(response.Licenses))
		}
	})
}

func TestLicenseHandler_GetDetails(t *testing.T) {
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
	handler := NewLicenseHandler(config, db)
	ctx := context.Background()
	logger := ctxzap.Extract(ctx).Sugar()

	t.Run("successful middleware processing", func(t *testing.T) {
		mockMW := &mockLicenseDetailsMiddleware{
			processFunc: func() (dto.LicenseRequestDTO, error) {
				return dto.LicenseRequestDTO{
					ID: "MIT",
				}, nil
			},
		}

		response, err := handler.GetDetails(ctx, logger, mockMW)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if response == nil {
			t.Fatal("Expected response, got nil")
		}

		if response.Status.Status != common.StatusCode_SUCCESS {
			t.Errorf("Expected SUCCESS status, got %v", response.Status.Status)
		}

		if response.License == nil {
			t.Error("Expected license to be initialized")
		}
	})

	t.Run("middleware processing error", func(t *testing.T) {
		mockMW := &mockLicenseDetailsMiddleware{
			processFunc: func() (dto.LicenseRequestDTO, error) {
				return dto.LicenseRequestDTO{}, errors.New("middleware error")
			},
		}

		response, err := handler.GetDetails(ctx, logger, mockMW)

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
