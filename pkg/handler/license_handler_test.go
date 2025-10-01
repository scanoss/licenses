package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
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
	processFunc func() ([]dto.ComponentRequestDTO, error)
}

func (m *mockMiddleware) Process() ([]dto.ComponentRequestDTO, error) {
	return m.processFunc()
}

type mockComponentMiddleware struct {
	processFunc func() (dto.ComponentRequestDTO, error)
}

func (m *mockComponentMiddleware) Process() (dto.ComponentRequestDTO, error) {
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
	handler := NewLicenseHandler(config, db)
	t.Run("successful middleware processing", func(t *testing.T) {
		mockMW := &mockMiddleware{
			processFunc: func() ([]dto.ComponentRequestDTO, error) {
				return []dto.ComponentRequestDTO{
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

	handler := NewLicenseHandler(config, db)

	t.Run("successful middleware processing", func(t *testing.T) {
		mockMW := &mockComponentMiddleware{
			processFunc: func() (dto.ComponentRequestDTO, error) {
				return dto.ComponentRequestDTO{
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
			processFunc: func() (dto.ComponentRequestDTO, error) {
				return dto.ComponentRequestDTO{}, errors.New("middleware error")
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
	handler := NewLicenseHandler(config, db)
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
