package handler

import (
	"context"
	"fmt"

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
	"scanoss.com/licenses/pkg/protocol/rest"
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
	s.Debugf("statusResp: %v", &statusResp)
	return &statusResp
}

func (h *LicenseHandler) GetComponentLicense(ctx context.Context, middleware middleware.Middleware[dto.ComponentRequestDTO]) (*pb.ComponentLicenseResponse, error) {
	s := ctxzap.Extract(ctx).Sugar()

	componentDTO, err := middleware.Process()
	if err != nil {
		return &pb.ComponentLicenseResponse{
			Status:    h.getResponseStatus(s, ctx, common.StatusCode_FAILED, rest.HTTP_CODE_400, err),
			Component: &pb.ComponentLicenseInfo{},
		}, nil
	}

	component, useCaseErr := h.licenseUseCase.GetComponentLicense(ctx, componentDTO)
	if useCaseErr != nil {
		return &pb.ComponentLicenseResponse{
			Status:    h.getResponseStatus(s, ctx, useCaseErr.Status, useCaseErr.Code, useCaseErr.Error),
			Component: &pb.ComponentLicenseInfo{},
		}, nil
	}

	return &pb.ComponentLicenseResponse{
		Status:    h.getResponseStatus(s, ctx, common.StatusCode_SUCCESS, rest.HTTP_CODE_200, nil),
		Component: component,
	}, nil
}

func (h *LicenseHandler) GetComponentsLicense(ctx context.Context, middleware middleware.Middleware[[]dto.ComponentRequestDTO]) (*pb.ComponentsLicenseResponse, error) {
	s := ctxzap.Extract(ctx).Sugar()

	componentsDTO, err := middleware.Process()
	if err != nil {
		return &pb.ComponentsLicenseResponse{
			Status:     h.getResponseStatus(s, ctx, common.StatusCode_FAILED, rest.HTTP_CODE_400, err),
			Components: []*pb.ComponentLicenseInfo{},
		}, nil
	}

	licenses, _ := h.licenseUseCase.GetComponentsLicense(ctx, componentsDTO)

	return &pb.ComponentsLicenseResponse{
		Status:     h.getResponseStatus(s, ctx, common.StatusCode_SUCCESS, rest.HTTP_CODE_200, err),
		Components: licenses,
	}, nil
}

func (h *LicenseHandler) GetDetails(ctx context.Context, middleware middleware.Middleware[dto.LicenseRequestDTO]) (*pb.LicenseDetailsResponse, error) {
	fmt.Print(ctx)
	s := ctxzap.Extract(ctx).Sugar()
	licenseDTO, err := middleware.Process()
	if err != nil {
		return &pb.LicenseDetailsResponse{
			Status:  h.getResponseStatus(s, ctx, common.StatusCode_FAILED, rest.HTTP_CODE_400, err),
			License: &pb.LicenseDetails{},
		}, err
	}
	licenseDetail, dErr := h.licenseUseCase.GetDetails(ctx, s, licenseDTO)

	if dErr != nil {
		s.Errorf("Error getting license details: %v", dErr)
		return &pb.LicenseDetailsResponse{
			Status:  h.getResponseStatus(s, ctx, dErr.Status, dErr.Code, dErr.Error),
			License: &pb.LicenseDetails{},
		}, dErr.Error
	}

	return &pb.LicenseDetailsResponse{
		Status:  h.getResponseStatus(s, ctx, common.StatusCode_SUCCESS, rest.HTTP_CODE_200, err),
		License: &licenseDetail,
	}, err

}
