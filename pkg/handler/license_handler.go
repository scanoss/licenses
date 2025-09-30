package handler

import (
	"context"
	"fmt"
	"net/http"
	"scanoss.com/licenses/pkg/helpers"
	"strconv"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jmoiron/sqlx"
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

type LicenseHandler struct {
	config         *myconfig.ServerConfig
	licenseUseCase *usecase.LicenseUseCase
}

// NewLicenseHandler creates a new instance of License handler.
func NewLicenseHandler(config *myconfig.ServerConfig, db *sqlx.DB) *LicenseHandler {
	return &LicenseHandler{
		config:         config,
		licenseUseCase: usecase.NewLicenseUseCase(config, db),
	}
}

func (h *LicenseHandler) getResponseStatus(s *zap.SugaredLogger, ctx context.Context, gRPCStatusCode common.StatusCode,
	httpCode int, msg string, err error) *common.StatusResponse {
	code := strconv.Itoa(httpCode)
	errHTTPCode := grpc.SetTrailer(ctx, metadata.Pairs("x-http-code", code))
	if errHTTPCode != nil {
		s.Debugf("error setting x-http-code to trailer: %v\n", errHTTPCode)
	}
	message := msg
	if err != nil {
		message = err.Error()
	}
	s.Debugf(message)
	statusResp := common.StatusResponse{Status: gRPCStatusCode, Message: message}
	s.Debugf("statusResp: %v", &statusResp)
	return &statusResp
}

func (h *LicenseHandler) GetComponentLicense(ctx context.Context, middleware middleware.Middleware[dto.ComponentRequestDTO]) (*pb.ComponentLicenseResponse, error) {
	s := ctxzap.Extract(ctx).Sugar()

	componentDTO, err := middleware.Process()
	if err != nil {
		return &pb.ComponentLicenseResponse{
			Status:    h.getResponseStatus(s, ctx, common.StatusCode_FAILED, http.StatusBadRequest, "", err),
			Component: &pb.ComponentLicenseInfo{},
		}, nil
	}

	componentLicenses, ucErr := h.licenseUseCase.GetComponentLicense(ctx, componentDTO)
	if ucErr != nil {
		return &pb.ComponentLicenseResponse{
			Status:    h.getResponseStatus(s, ctx, ucErr.Status, ucErr.Code, "", ucErr.Error),
			Component: &pb.ComponentLicenseInfo{},
		}, nil
	}
	grpcStatus, httpCode, message := helpers.DetermineStatusResponse(componentLicenses)
	return &pb.ComponentLicenseResponse{
		Status:    h.getResponseStatus(s, ctx, grpcStatus, httpCode, message, nil),
		Component: componentLicenses,
	}, nil
}

func (h *LicenseHandler) GetComponentsLicense(ctx context.Context, middleware middleware.Middleware[[]dto.ComponentRequestDTO]) (*pb.ComponentsLicenseResponse, error) {
	s := ctxzap.Extract(ctx).Sugar()

	componentsDTO, err := middleware.Process()
	if err != nil {
		return &pb.ComponentsLicenseResponse{
			Status:     h.getResponseStatus(s, ctx, common.StatusCode_FAILED, http.StatusBadRequest, "", err),
			Components: []*pb.ComponentLicenseInfo{},
		}, nil
	}

	componentLicenses, ucErr := h.licenseUseCase.GetComponentsLicense(ctx, componentsDTO)
	if ucErr != nil {
		return &pb.ComponentsLicenseResponse{
			Status:     h.getResponseStatus(s, ctx, ucErr.Status, ucErr.Code, "", ucErr.Error),
			Components: []*pb.ComponentLicenseInfo{},
		}, nil
	}
	grpcStatus, httpCode, message := helpers.DetermineStatusResponse(componentLicenses)
	return &pb.ComponentsLicenseResponse{
		Status:     h.getResponseStatus(s, ctx, grpcStatus, httpCode, message, nil),
		Components: componentLicenses,
	}, nil
}

func (h *LicenseHandler) GetDetails(ctx context.Context, middleware middleware.Middleware[dto.LicenseRequestDTO]) (*pb.LicenseDetailsResponse, error) {
	fmt.Print(ctx)
	s := ctxzap.Extract(ctx).Sugar()
	licenseDTO, err := middleware.Process()
	if err != nil {
		return &pb.LicenseDetailsResponse{
			Status:  h.getResponseStatus(s, ctx, common.StatusCode_FAILED, http.StatusBadRequest, "", err),
			License: &pb.LicenseDetails{},
		}, err
	}
	licenseDetail, ucErr := h.licenseUseCase.GetDetails(ctx, s, licenseDTO)

	if ucErr != nil {
		s.Errorf("Error getting license details: %v", ucErr)
		return &pb.LicenseDetailsResponse{
			Status:  h.getResponseStatus(s, ctx, ucErr.Status, ucErr.Code, "", ucErr.Error),
			License: &pb.LicenseDetails{},
		}, nil
	}
	status, httpCode, message := helpers.DetermineStatusResponse(&licenseDetail)
	return &pb.LicenseDetailsResponse{
		Status:  h.getResponseStatus(s, ctx, status, httpCode, message, err),
		License: &licenseDetail,
	}, err

}
