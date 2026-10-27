package server

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/licensesv2"
	"scanoss.com/licenses/pkg/cache"
	myconfig "scanoss.com/licenses/pkg/config"
	"scanoss.com/licenses/pkg/handler"
	"scanoss.com/licenses/pkg/middleware"
)

type LicenseServer struct {
	// Embedding the generated Unimplemented struct, not the LicenseServer
	// interface. The interface value is nil, so any method the service does not
	// implement dereferences it and panics, taking the whole process down - and
	// GetObligations is unimplemented here (and Echo was too), so a single
	// unauthenticated call was enough. The struct answers those with an
	// Unimplemented error instead, and keeps doing so as papi grows new methods.
	pb.UnimplementedLicenseServer
	config  *myconfig.ServerConfig
	handler *handler.LicenseHandler
	db      *sqlx.DB
}

// NewLicenseServer creates a new instance of Licenses Server.
func NewLicenseServer(config *myconfig.ServerConfig, db *sqlx.DB, spdxCache cache.SPDXLicenseCacheInterface) pb.LicenseServer {
	return &LicenseServer{
		config:  config,
		db:      db,
		handler: handler.NewLicenseHandler(config, db, spdxCache),
	}
}

// Echo sends back the same message received, as every other service does. It was
// left unimplemented here, so the liveness check answered HTTP 500.
func (ls LicenseServer) Echo(_ context.Context, request *commonv2.EchoRequest) (*commonv2.EchoResponse, error) {
	return &commonv2.EchoResponse{Message: request.GetMessage()}, nil
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
