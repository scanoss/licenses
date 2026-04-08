# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
## [0.2.0] - 2026-04-10
### Added
- Added `LOOKUP_SOURCE_PRIORITY` configuration (env var and JSON `Lookup.SourcePriority`) to control the ordered priority of license detection sources. Sources are walked from highest to lowest priority, stopping at the first source that returns license data. See [README](README.md#license-lookup-source-priority) for details.
### Changed
- Cut license search on the first hit using the configured source priority

## [0.1.0] - 2026-04-07
### Added
- Added nearest version fallback for license lookup: when no licenses exist for a specific version, queries all known versions and returns licenses for the nearest version to the requirement
- Added `go-component-helper` dependency for centralized component version resolution
- Added per-component `error_message` and `error_code` fields in license responses for failed components
- Added SPDX license cache with configurable refresh interval (`CACHE_SPDX_REFRESH_HOURS`, default 24h)
- Added unit tests for SPDX license cache
- Added `AttributionFile` license source

### Changed
- Upgraded golangci-lint configuration to v2
- Fixed linter issues across the codebase
- Enhanced package documentation with `doc.go` files
- Added `full_name`, `url`, and `is_spdx_approved` fields to SPDX license response using cached license details
- Replaced local `dto.ComponentRequestDTO` with `componenthelper.ComponentDTO` from `go-component-helper`
- Replaced per-component version resolution with `componenthelper.GetComponentsVersion` using worker pool
- Removed `groupComponentsByPurl` deduplication logic from batch middleware
- Changed component license endpoints to always return `SUCCESS` overall status, with errors reported per-component via `error_message` and `error_code`
- Added unit tests for per-component error response status
- Updated dependencies to the latest versions

### Removed
- Removed local `replace` directives for `go-component-helper` and `go-models`

## [0.0.7] - 2025-10-01
### Added
- Added `response_helper.go` with status determination logic for license responses

### Changed
- Refactored license handler to use centralized status response logic via `DetermineStatusResponse`
- Updated HTTP status code handling to use `int` type instead of string constants
- Improved error response messages for component license lookups
- Enhanced response status logic to handle warnings when some components have no licenses

### Removed
- Removed `pkg/protocol/rest/http_code.go` in favor of standard `net/http` status codes


## [0.0.6] - 2025-09-30
### Updated
- Updated papi dependency to `v0.24.0`
### Changed
- Changed default ports: REST `40057`, gRPC `50057`, and logging `66057`


[0.2.0]: https://github.com/scanoss/licenses/releases/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/scanoss/licenses/releases/compare/v0.0.7...v0.1.0
[0.0.7]: https://github.com/scanoss/licenses/releases/compare/v0.0.6...v0.0.7
[0.0.6]: https://github.com/scanoss/licenses/releases/compare/v0.0.1...v0.0.6
[0.0.1]: https://github.com/scanoss/licenses/releases/tag/v0.0.1
