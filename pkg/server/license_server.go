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

// NewLicenseServer creates a new instance of Licenses Server.
func NewLicenseServer(config *myconfig.ServerConfig, db *sqlx.DB) pb.LicenseServer {
	return &LicenseServer{
		config:  config,
		db:      db,
		handler: handler.NewLicenseHandler(config, db),
	}
}

// GetComponentLicenses search licenses for one component.
func (ls LicenseServer) GetComponentLicenses(ctx context.Context, req *commonv2.ComponentRequest) (*pb.ComponentLicenseResponse, error) {
	return ls.handler.GetComponentLicense(ctx, middleware.NewComponentRequestMiddleware(req, ctx))
}

// GetComponentsLicenses search licenses for multiple components in a single request.
func (ls LicenseServer) GetComponentsLicenses(ctx context.Context, request *commonv2.ComponentsRequest) (*pb.ComponentsLicenseResponse, error) {
	return ls.handler.GetComponentsLicense(ctx, middleware.NewComponentsRequestMiddleware(request, ctx))
}

// GetDetails searches for license information.
func (ls LicenseServer) GetDetails(ctx context.Context, request *pb.LicenseRequest) (*pb.LicenseDetailsResponse, error) {
	return ls.handler.GetDetails(ctx, middleware.NewLicenseDetailMiddleware(request, ctx))
}
