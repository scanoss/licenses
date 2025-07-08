package handler

import (
	"context"
	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/licensesv2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	myconfig "scanoss.com/licenses/pkg/config"
	"scanoss.com/licenses/pkg/dto"
	"scanoss.com/licenses/pkg/middleware"
	"scanoss.com/licenses/pkg/usecase"
)

const (
	// Código HTTP
	HTTP_CODE_200 = "200"
	HTTP_CODE_400 = "400"
	HTTP_CODE_500 = "500"
)

type LicenseHandler struct {
	config *myconfig.ServerConfig
}

// NewLicenseHandler creates a new instance of License handler.
func NewLicenseHandler(config *myconfig.ServerConfig) *LicenseHandler {
	return &LicenseHandler{config: config}
}

func (h *LicenseHandler) getResponseStatus(s *zap.SugaredLogger, ctx context.Context, gRPCStatusCode common.StatusCode,
	httpCode string, err error) *common.StatusResponse {
	errHTTPCode := grpc.SetTrailer(ctx, metadata.Pairs("x-http-code", httpCode))
	if errHTTPCode != nil {
		s.Debugf("error setting x-http-code to trailer: %v\n", err)
	}
	message := "Licenses Successfully retrieved"
	if err != nil {
		message = err.Error()
	}
	s.Debugf(message)
	statusResp := common.StatusResponse{Status: gRPCStatusCode, Message: message}
	return &statusResp
}

func (h *LicenseHandler) GetLicenses(ctx context.Context, s *zap.SugaredLogger,
	middleware middleware.Middleware[[]dto.ComponentRequestDTO]) (*pb.BasicResponse, error) {
	componentsDTO, err := middleware.Process()
	if err != nil {
		return &pb.BasicResponse{
			Status:   h.getResponseStatus(s, ctx, common.StatusCode_FAILED, HTTP_CODE_400, err),
			Licenses: make([]*pb.BasicLicenseResponse, 0)}, err
	}
	lu := usecase.NewLicenseUseCase(s, h.config)
	lu.GetLicenses(ctx, componentsDTO)
	return &pb.BasicResponse{
		Status:   h.getResponseStatus(s, ctx, common.StatusCode_SUCCESS, HTTP_CODE_200, err),
		Licenses: make([]*pb.BasicLicenseResponse, 0),
	}, err
}

func (h *LicenseHandler) GetDetails(ctx context.Context, s *zap.SugaredLogger,
	middleware middleware.Middleware[dto.LicenseRequestDTO]) (*pb.DetailsResponse, error) {
	licenseDTO, err := middleware.Process()
	if err != nil {
		return &pb.DetailsResponse{
			Status:  h.getResponseStatus(s, ctx, common.StatusCode_FAILED, HTTP_CODE_400, err),
			License: &pb.LicenseResponse{},
		}, err
	}
	lu := usecase.NewLicenseUseCase(s, h.config)
	lu.GetDetails(ctx, licenseDTO)
	return &pb.DetailsResponse{
		Status:  h.getResponseStatus(s, ctx, common.StatusCode_SUCCESS, HTTP_CODE_200, err),
		License: &pb.LicenseResponse{},
	}, err

}
