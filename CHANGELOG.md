# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Upcoming changes...

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

## [0.0.1] - 2022-04-22

[0.0.7]: https://github.com/scanoss/licenses/releases/tag/v0.0.7
[0.0.6]: https://github.com/scanoss/licenses/releases/tag/v0.0.6
[0.0.1]: https://github.com/scanoss/licenses/releases/tag/v0.0.1
