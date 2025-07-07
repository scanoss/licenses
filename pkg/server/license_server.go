package server

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
	"github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/licensesv2"
	myconfig "scanoss.com/licenses/pkg/config"
	"scanoss.com/licenses/pkg/handler"
	"scanoss.com/licenses/pkg/middleware"
)

type LicenseServer struct {
	pb.LicenseServer
	config  *myconfig.ServerConfig
	handler *handler.LicenseHandler
	db      *sqlx.DB
}

// NewLicenseServer creates a new instance of License Server.
func NewLicenseServer(config *myconfig.ServerConfig, db *sqlx.DB) pb.LicenseServer {
	return &LicenseServer{
		config:  config,
		db:      db,
		handler: handler.NewLicenseHandler(config),
	}
}

// GetLicenses searches for license information.
func (pb LicenseServer) GetLicenses(ctx context.Context, request *commonv2.ComponentBatchRequest) (*pb.BasicResponse, error) {
	s := ctxzap.Extract(ctx).Sugar()
	return pb.handler.GetLicenses(ctx, s, middleware.NewComponentBatchMiddleware(request, s))
}

// GetDetails searches for license information.
func (pb LicenseServer) GetDetails(ctx context.Context, request *pb.LicenseRequest) (*pb.DetailsResponse, error) {
	s := ctxzap.Extract(ctx).Sugar()
	return pb.handler.GetDetails(ctx, s, middleware.NewLicenseDetailMiddleware(request, s))
}
