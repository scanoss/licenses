package middleware

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/scanoss/papi/api/commonv2"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	"scanoss.com/licenses/pkg/dto"

	"testing"
)

func TestComponentBatchMiddleware(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()
	middleware := &ComponentBatchMiddleware[[]dto.ComponentRequestDTO]{
		MiddlewareBase: MiddlewareBase{s: s},
		req: &commonv2.ComponentBatchRequest{
			Components: []*commonv2.ComponentRequest{
				{Purl: "pkg:npm/lodash", Requirement: "4.17.21"},
				{Purl: "pkg:npm/react", Requirement: "18.0.0"},
				{Purl: "pkg:npm/react"},
				{Purl: "pkg:npm/react", Requirement: "18.0.0"},
				{Purl: "pkg:github/scanoss/scanenr.c@1.2.3"},
				{Purl: "pkg:github/scanoss/scanenr.c@1.2.3"},
			},
		},
	}

	t.Run("should process requests", func(t *testing.T) {
		result, err := middleware.Process()
		if err != nil {
			t.Fatalf("Failed to process requests. Expected 4 components, got %d", len(result))
		}
	})
}

func TestNewComponentBatchMiddlewareWithEmptyRequest(t *testing.T) {
	ctx := context.Background()
	s := ctxzap.Extract(ctx).Sugar()
	middleware := &ComponentBatchMiddleware[[]dto.ComponentRequestDTO]{
		MiddlewareBase: MiddlewareBase{s: s},
		req: &commonv2.ComponentBatchRequest{
			Components: []*commonv2.ComponentRequest{},
		},
	}

	t.Run("should not process empty requests", func(t *testing.T) {
		components, err := middleware.Process()
		if err == nil {
			t.Fatalf("Should not process empty requests, but got %d components", len(components))
		}
	})
}

func TestGroupComponentsByPurl(t *testing.T) {
	ctx := context.Background()
	s := ctxzap.Extract(ctx).Sugar()
	middleware := &ComponentBatchMiddleware[[]dto.ComponentRequestDTO]{
		MiddlewareBase: MiddlewareBase{s: s},
	}

	t.Run("component grouping", func(t *testing.T) {
		components := []dto.ComponentRequestDTO{
			{Purl: "pkg:npm/lodash", Requirement: "4.17.21"},
			{Purl: "pkg:npm/react", Requirement: "18.0.0"},
			{Purl: "pkg:npm/react"},
			{Purl: "pkg:npm/react", Requirement: "18.0.0"},
			{Purl: "pkg:github/scanoss/scanenr.c@1.2.3"},
			{Purl: "pkg:github/scanoss/scanenr.c@1.2.3"},
		}

		result := middleware.groupComponentsByPurl(components)

		if len(result) != 4 {
			t.Fatalf("Failed to group components by purl. Expected 4 components, got %d", len(result))
		}

	})
}
