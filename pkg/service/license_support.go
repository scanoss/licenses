// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2025 SCANOSS.COM
 */

package service

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"scanoss.com/licenses/pkg/config"
)

type LicenseMetrics struct {
	config *config.ServerConfig
	meter  metric.Meter

	// Request-level metrics
	licenseRequestsTotal         metric.Int64Counter     // Total count of all license API requests
	licenseRequestDuration       metric.Float64Histogram // Time taken to process each request (in seconds)
	licenseRequestComponentCount metric.Int64Histogram   // Number of components/PURLs per request

	// Granular business metrics for PURL processing
	purlMalformedTotal            metric.Int64Counter // Invalid or unparseable PURL format
	purlComponentNotFoundTotal    metric.Int64Counter // Component/package not found in database
	purlNoLicenseVersionedTotal   metric.Int64Counter // No license found for the specific version requested
	purlLicenseFoundExactTotal    metric.Int64Counter // License found for the exact version requested
	purlLicenseFoundFallbackTotal metric.Int64Counter // License found using unversioned (fallback) lookup
	purlNoLicenseFallbackTotal    metric.Int64Counter // No license found even with fallback lookup

	// License processing metrics
	licenseParsingTotal metric.Int64Counter // Total SPDX license expression parsing attempts

}

func NewLicenseMetrics(config *config.ServerConfig) (*LicenseMetrics, error) {
	if !config.Telemetry.Enabled {
		return &LicenseMetrics{config: config}, nil
	}

	meter := otel.Meter("scanoss.com/licenses")

	lm := &LicenseMetrics{
		config: config,
		meter:  meter,
	}

	var err error

	// Request-level metrics
	lm.licenseRequestsTotal, err = meter.Int64Counter("license_requests_total", metric.WithDescription("Total number of license requests processed"))
	if err != nil {
		return nil, err
	}

	lm.licenseRequestDuration, err = meter.Float64Histogram("license_request_duration_seconds", metric.WithDescription("Duration of license request processing"), metric.WithUnit("s"))
	if err != nil {
		return nil, err
	}

	lm.licenseRequestComponentCount, err = meter.Int64Histogram("license_request_component_count", metric.WithDescription("Number of components per license request"))
	if err != nil {
		return nil, err
	}

	// Granular business metrics for quality tracking
	lm.purlMalformedTotal, err = meter.Int64Counter("purl_malformed_total", metric.WithDescription("Total number of malformed or invalid PURLs"))
	if err != nil {
		return nil, err
	}

	lm.purlComponentNotFoundTotal, err = meter.Int64Counter("purl_component_not_found_total", metric.WithDescription("Total number of PURLs where component version could not be resolved"))
	if err != nil {
		return nil, err
	}

	lm.purlNoLicenseVersionedTotal, err = meter.Int64Counter("purl_no_license_versioned_total", metric.WithDescription("Total number of PURLs where no license was found for the specific version"))
	if err != nil {
		return nil, err
	}

	lm.purlLicenseFoundExactTotal, err = meter.Int64Counter("purl_license_found_exact_total", metric.WithDescription("Total number of PURLs where licenses were found with exact version match"))
	if err != nil {
		return nil, err
	}

	lm.purlLicenseFoundFallbackTotal, err = meter.Int64Counter("purl_license_found_fallback_total", metric.WithDescription("Total number of PURLs where licenses were found using fallback (unversioned) lookup"))
	if err != nil {
		return nil, err
	}

	lm.purlNoLicenseFallbackTotal, err = meter.Int64Counter("purl_no_license_fallback_total", metric.WithDescription("Total number of PURLs where no license was found even in fallback (unversioned) lookup"))
	if err != nil {
		return nil, err
	}

	// License processing metrics
	lm.licenseParsingTotal, err = meter.Int64Counter("license_parsing_total", metric.WithDescription("Total number of license expression parsing operations"))
	if err != nil {
		return nil, err
	}

	return lm, nil
}

// RecordLicenseRequest records metrics for a license request
func (lm *LicenseMetrics) RecordLicenseRequest(ctx context.Context, componentCount int, duration time.Duration, success bool, requestType string) {
	if !lm.config.Telemetry.Enabled {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("request_type", requestType),
		attribute.Bool("success", success),
	}

	lm.licenseRequestsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
	lm.licenseRequestDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	lm.licenseRequestComponentCount.Record(ctx, int64(componentCount), metric.WithAttributes(attrs...))
}

// RecordPURLComponentNotFound records when a component version cannot be resolved
func (lm *LicenseMetrics) RecordPURLComponentNotFound(ctx context.Context, purl string) {
	if !lm.config.Telemetry.Enabled {
		return
	}

	lm.purlComponentNotFoundTotal.Add(ctx, 1)
}

// RecordPURLNoLicenseVersioned records when no license is found for a specific version
func (lm *LicenseMetrics) RecordPURLNoLicenseVersioned(ctx context.Context, purl string) {
	if !lm.config.Telemetry.Enabled {
		return
	}

	lm.purlNoLicenseVersionedTotal.Add(ctx, 1)
}

// RecordPURLLicenseFoundExact records when licenses are found with exact version match
func (lm *LicenseMetrics) RecordPURLLicenseFoundExact(ctx context.Context, purl string) {
	if !lm.config.Telemetry.Enabled {
		return
	}

	lm.purlLicenseFoundExactTotal.Add(ctx, 1)
}

// RecordPURLLicenseFoundFallback records when licenses are found using fallback (unversioned) lookup
func (lm *LicenseMetrics) RecordPURLLicenseFoundFallback(ctx context.Context, purl string) {
	if !lm.config.Telemetry.Enabled {
		return
	}

	lm.purlLicenseFoundFallbackTotal.Add(ctx, 1)
}

// RecordPURLNoLicenseFallback records when no license is found even in fallback lookup
func (lm *LicenseMetrics) RecordPURLNoLicenseFallback(ctx context.Context, purl string) {
	if !lm.config.Telemetry.Enabled {
		return
	}

	lm.purlNoLicenseFallbackTotal.Add(ctx, 1)
}

// RecordLicenseParsingResult records license expression parsing results
func (lm *LicenseMetrics) RecordLicenseParsingResult(ctx context.Context, success bool) {
	if !lm.config.Telemetry.Enabled || lm.licenseParsingTotal == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.Bool("success", success),
	}

	lm.licenseParsingTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
}
