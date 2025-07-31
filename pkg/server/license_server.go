package server

import (
	"context"

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
		handler: handler.NewLicenseHandler(config, db),
	}
}

// GetLicenses searches for license information.
func (ls LicenseServer) GetLicenses(ctx context.Context, request *commonv2.ComponentBatchRequest) (*pb.BatchLicenseResponse, error) {
	return ls.handler.GetLicenses(ctx, middleware.NewComponentBatchMiddleware(request, ctx))
}

// GetDetails searches for license information.
func (ls LicenseServer) GetDetails(ctx context.Context, request *pb.LicenseRequest) (*pb.LicenseDetailsResponse, error) {
	return ls.handler.GetDetails(ctx, middleware.NewLicenseDetailMiddleware(request, ctx))
}
