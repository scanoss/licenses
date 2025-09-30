package helpers

import (
	"fmt"
	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/licensesv2"
	"net/http"
)

// componentsLicenseInfoStatus determines the appropriate status code and message
// for a list of component license information.
//
// It analyzes the list to determine if all, some, or none of the components have
// associated licenses, and returns the corresponding gRPC status code, HTTP status code,
// and descriptive message.
//
// Returns:
//   - FAILED with 404 if the input is nil or no licenses found for any component
//   - SUCCEEDED_WITH_WARNINGS with 200 if some components have licenses and some don't
//   - SUCCESS with 200 if all components have licenses
func componentsLicenseInfoStatus(componentLicenses []*pb.ComponentLicenseInfo) (gRPCStatusCode common.StatusCode,
	httpCode int, message string) {
	if componentLicenses == nil {
		return common.StatusCode_FAILED, http.StatusNotFound, "Licenses not found for requested component(s)"
	}
	var notFound []string
	success := 0
	for _, componentLicense := range componentLicenses {
		if len(componentLicense.Licenses) == 0 {
			notFound = append(notFound, componentLicense.Purl)
			continue
		}
		success++
	}
	if success == 0 {
		return common.StatusCode_FAILED, http.StatusNotFound, fmt.Sprintf("No licenses found for the following component(s):%v", notFound)
	}

	if success > 0 && len(notFound) > 0 {
		return common.StatusCode_SUCCEEDED_WITH_WARNINGS, http.StatusOK, fmt.Sprintf("No licenses found for the following component(s):%v", notFound)
	}

	return common.StatusCode_SUCCESS, http.StatusOK, "Licenses retrieved successfully"
}

// componentLicenseInfoStatus determines the appropriate status code and message
// for a single component license information.
//
// Returns:
//   - FAILED with 404 if the input is nil or no licenses found
//   - SUCCESS with 200 if licenses are found
func componentLicenseInfoStatus(componentLicenses *pb.ComponentLicenseInfo) (gRPCStatusCode common.StatusCode,
	httpCode int, message string) {
	if componentLicenses == nil {
		return common.StatusCode_FAILED, http.StatusNotFound, "Licenses not found for requested component"
	}
	return componentsLicenseInfoStatus([]*pb.ComponentLicenseInfo{componentLicenses})
}

// licenseDetailsStatus determines the appropriate status code and message
// for license details information.
//
// Returns:
//   - FAILED with 404 if the input is nil
//   - SUCCESS with 200 if license details are found
func licenseDetailsStatus(licenseDetails *pb.LicenseDetails) (gRPCStatusCode common.StatusCode,
	httpCode int, message string) {
	if licenseDetails == nil {
		return common.StatusCode_FAILED, http.StatusNotFound, "License details not found"
	}
	return common.StatusCode_SUCCESS, http.StatusOK, "License details retrieved successfully"
}

// DetermineStatusResponse is the main entry point for determining appropriate status
// codes and messages based on the type and content of the response data.
//
// It performs type assertion on the input data and delegates to the appropriate
// specific status function based on the data type.
//
// Supported types:
//   - []*pb.ComponentLicenseInfo: List of component license information
//   - *pb.ComponentLicenseInfo: Single component license information
//   - *pb.LicenseDetails: License details information
//
// Returns:
//   - gRPCStatusCode: The gRPC status code (SUCCESS, SUCCEEDED_WITH_WARNINGS, or FAILED)
//   - httpCode: The corresponding HTTP status code (200 or 404)
//   - message: A descriptive message about the operation result
//   - For unsupported types, returns FAILED with 500 Internal Server Error
func DetermineStatusResponse(data interface{}) (gRPCStatusCode common.StatusCode,
	httpCode int, message string) {
	switch v := data.(type) {
	case []*pb.ComponentLicenseInfo:
		return componentsLicenseInfoStatus(v)
	case *pb.ComponentLicenseInfo:
		return componentLicenseInfoStatus(v)
	case *pb.LicenseDetails:
		return licenseDetailsStatus(v)
	default:
		return common.StatusCode_FAILED, http.StatusInternalServerError, "Internal server error"
	}
}
