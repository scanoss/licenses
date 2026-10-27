# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.5.2] - 2026-10-27

### Fixed
- **The release workflow fails to build.** Since 0.5.1 it runs in a `golang:1.25-bookworm` container as root, over a checkout the runner owns, and git refuses such a repository, so the build could not read the version from the tag and the build stopped. The workflow now marks the workspace as a safe directory. 0.5.1 was tagged but never released

## [0.5.1] - 2026-10-27

### Fixed
- **The 0.5.0 binaries and Docker image do not start on Debian bookworm.** The DuckDB driver made the build cgo, so the binary links against glibc, and both builds ran on a newer one than the bookworm runtime: the release on `ubuntu-latest` and the image on `golang:1.25`, now Debian 13. They require `GLIBC_2.38` and bookworm has 2.36, so they failed at startup with "version `GLIBC_2.38' not found". The release now builds in a `golang:1.25-bookworm` container and the Dockerfile's build stage uses the same image; the binaries require glibc 2.34 at most

## [0.5.0] - 2026-10-27

### Changed
- **The service now reads the serving knowledge base instead of the mining database.** The mining path is gone; there is no configuration that selects it. Endpoint routes and response shapes are unchanged
- **Licences are keyed by `(purl_id, version)`.** The mining `purl_licenses` table carried the purl string on every row; serving records them in `component_version_license`, keyed by the component's UUID, so every lookup now takes a resolved `purl_id`. The purl is joined back into the result, so callers see what they saw before
- **OSADL's three compatibility lists are read from their own tables.** Mining kept `compatibilities`, `incompatibilities` and `depending_compatibilities` as PostgreSQL arrays on the `osadl` row; serving normalises each into `OSADL_compat`, `OSADL_incompat` and `OSADL_depending`, which are read and reassembled into the same response
- **The SPDX registry is `"SPDX_license_data"`** with renamed columns (`SPDX_id`, `is_deprecated`, `details_url`, `reference_number`, `is_osi_approved`), aliased back in the projection. The same 825 rows
- Component and version resolution no longer goes through `go-component-helper`, which resolves against `all_urls`, `mines` and `versions` - none of which exist in the serving schema. It is replaced by `pkg/resolver`
- Whether a licence is SPDX-approved is no longer a stored flag: mining's `licenses.is_spdx` has no serving equivalent, so it is derived the way the rest of the schema expresses it - the identifier appears in the SPDX registry
- Binaries are built with `CGO_ENABLED=1`, and the runtime image moved to `debian:bookworm-slim` with `libstdc++6`. The DuckDB driver is a cgo binding, so a pure-Go build can no longer be produced
- The licence lookups no longer join `component`. They did only to return a purl nothing read, and the unindexed join cost 2.7 s of CPU per lookup against 5 ms without it
- A source link that points back at the component itself (11.4 of 13.9 million) is no longer followed. It cost an unindexed full scan of `component`, about 2.5 s of CPU, for a fallback that could only repeat the lookup

### Added
- **A package-level licence fallback.** When no release of a component has a licence, the lookup falls to `component_license`, the export's per-component licence table. `component_version_license` covers 22,849,935 components and `component_license` covers all 31,761,182, so 8,911,247 components - 28% - have a licence recorded there and nowhere else, and used to be answered as having none. Its rows carry an empty version, because the licence genuinely is not attributed to a release; reporting the requested version would invent an attribution the database does not make. It runs only when `Lookup.SourcePriority` includes source 0, since everything `component_license` records is declared
- **DuckDB support**, the intended production backend for the serving knowledge base. Set `DB_DRIVER=duckdb` and `DB_PATH` to the serving file, which is opened read-only
- `pkg/db`, opening the serving database for any of duckdb, sqlite or postgres
- `pkg/resolver`, the serving component resolution
- `pkg/model/license_registry.go`, replacing two go-models calls that cannot work against serving: both query the mining `licenses` and `spdx_license_data` tables, which serving renames with different columns. The go-models result types are reused, so the SPDX cache and the licence lookup are unchanged apart from where they read from
- `pkg/model/test_schema.go`, holding the serving schema the tests build their database from. The fixtures under `tests/` used to carry their own `CREATE TABLE` statements - the `osadl` one declared columns the production table does not have

### Security
- **An unauthenticated request could kill the server.** `LicenseServer` embedded the papi `LicenseServer` *interface* rather than the generated `UnimplementedLicenseServer`, and that field is nil, so any method the service does not implement dereferenced it and panicked, taking the whole process down. Two of the five - `Echo` and `GetObligations` - are unimplemented, so `POST /v2/licenses/echo` or `GET /v2/licenses/obligations` was enough. The generated struct answers those with an `Unimplemented` error instead. Pre-existing and unrelated to the serving migration
- Upgraded `google.golang.org/grpc` to v1.83.1, `golang.org/x/net` to v0.57.0 and the OpenTelemetry modules (`otel`, `otel/sdk`, `otlptrace`, `otlptracegrpc`) to v1.45.0 to fix open Dependabot security advisories
- `go-grpc-helper` v0.16.0 → v0.17.0. A failed REST call no longer answers `200` with an empty, unparseable body when the handler set no `x-http-code` trailer: the gateway now falls back to the standard grpc-gateway error response, and gRPC status codes map to their HTTP equivalent, so `Unimplemented` is `501` rather than `500`

### Removed
- `pkg/model/ldb_component_licenses.go`. It queried the mining `ldb_component_licenses` table and nothing instantiated it
- `date` is gone from the licence rows. The serving schema records none, so the field comes back empty; nothing read it

### Deployment
- **Breaking:** requires a serving knowledge base. Point `DB_PATH` at the `.duckdb` file and set `DB_DRIVER=duckdb`
- **Serving carries the declared licence only.** `component_version_license.source` is `0` on all 205,234,732 rows, where mining's `purl_licenses` spread 374,115,789 rows across eight sources (0: 274M, 5: 53M, 3: 18M, 31: 17M, 33: 9.9M, plus 6/32/34). The detected licences from the other sources are deliberately out of scope, so a single source is the expected shape and `Lookup.SourcePriority` resolves on `0`. The filtering is kept as written rather than removed, so nothing has to be reintroduced if the export ever carries more sources
- **Coverage checked per component**, not per version string: of 40 randomly sampled components that mining licenses under source 0, **36 have at least as many licensed versions in serving**, 2 have one fewer, and 2 have none - both `pkg:apk/*`, which exist in `component` but carry no `component_version` rows at all, so that is a version-coverage gap rather than a licence one
- **Version strings differ between the two schemas, and a naive comparison will look like catastrophic loss.** Serving normalises them: mining's `v0.1.1` is `0.1.1`, and tag-shaped versions (`datahub-0.2.188`, `node-http2-0.0.6`, `OpenSSL_1_1_1a`) are rewritten entirely. Joining mining and serving on the exact version string reports most releases as missing when the component and its releases are both present under different spellings. Compare per component
- `OSADL` has 121 rows against mining's 124
- **The export added tables that are still empty:** `detected_license`, `notices`, `LicenseComplexExpression` and `component_version_attribution` all exist with 0 rows. `OSADL_usecase` (219 rows) is populated and has no field in the papi contract
- Verified against the production serving file (72.6 GiB): per-component licences, the batch endpoint, SPDX details including the `see_also` array and the reassembled OSADL block, and the 825-license SPDX cache. The package-level fallback was re-verified against the updated file: `pkg:github/commonroad/sandra` has no per-version licence and is answered from `component_license`, while `pkg:npm/express` keeps its per-release ones
- The Dockerfile was already broken before this change - it built `./cmd/server`, which does not exist, and named the binary `scanoss-dependencies`. It now builds `./cmd/licenses-api` as `scanoss-licenses`
- `make build_arm` now cross-compiles with cgo and needs an `aarch64` toolchain (set `CC`); the amd64 build is unaffected

### Fixed
- **Scoped npm packages, and other purls the export stores encoded or with qualifiers, resolve again.** `component.purl` keeps scoped npm and jsr names percent-encoded (`pkg:npm/%40angular/core`), qualifiers on generic, conda and cran keys, and a raw `+` in some rpm and deb names. The resolver decoded the purl and rebuilt the key, so 1.96 million components (6%) - plus 1.6 million qualified ones - answered `COMPONENT_NOT_FOUND` however they were spelled. It now tries the caller's spelling, the canonical rendering and the decoded form in one indexed lookup
- **`Echo` is implemented.** It was the only service without it, so the echo endpoint answered HTTP 500
- **The package-level licence fallback is skipped when the export has no `component_license`.** The 2026-10-16 export dropped it, and every lookup that reached the fallback logged a Catalog Error. The model probes for the table once; any other failure is retried

## [0.4.0] - 2026-06-22
### Added
- `/health` liveness endpoint (GET) on the REST gateway
### Changed
- Upgraded `scanoss/go-grpc-helper` to `v0.16.0`

## [0.3.1] - 2026-05-07
- Fixed `licenses/details` endpoint

## [0.3.0] - 2026-04-20
### Fixed
- Fixed `is_spdx_approved` being set to `true` for licenses whose details could not be found in the SPDX cache; the flag is now only set when the cache actually returns details.
### Changed
- Upgrade `scanoss/grpc-helper` to `v0.15.1`
- Upgraded `scanoss/papi` to `v0.37.0`.
- Changed the per-component `error_code` field from `pb.StatusCode` protobuf enum to a plain `string`.
- Replace `error_message` and `error_code` with `info_message` and `info_code`
- Renamed error code `ComponentWithoutInfo` to `NoInfo`.
- Added new status `RequirementNotMet` to response

## [0.2.1] - 2026-04-14
### Fixed
- Fixed `GetComponentsLicense` dropping components with resolution errors from the response; failed components are now included alongside successful lookup results.

## [0.2.0] - 2026-04-10
### Added
- Added `LOOKUP_SOURCE_PRIORITY` configuration (env var and JSON `Lookup.SourcePriority`) to control the ordered priority of license detection sources. Sources are walked from highest to lowest priority, stopping at the first source that returns license data. See [README](README.md#license-lookup-source-priority) for details.
- Added concurrent license lookup with bounded concurrency and context cancellation support. The number of in-flight license lookups is configurable via `LOOKUP_MAX_WORKERS` (env var and JSON `Lookup.MaxWorkers`).
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


[0.5.2]: https://github.com/scanoss/licenses/compare/v0.5.1...v0.5.2
[0.5.1]: https://github.com/scanoss/licenses/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/scanoss/licenses/compare/v0.4.0...v0.5.0
[0.3.1]: https://github.com/scanoss/licenses/releases/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/scanoss/licenses/releases/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/scanoss/licenses/releases/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/scanoss/licenses/releases/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/scanoss/licenses/releases/compare/v0.0.7...v0.1.0
[0.0.7]: https://github.com/scanoss/licenses/releases/compare/v0.0.6...v0.0.7
[0.0.6]: https://github.com/scanoss/licenses/releases/compare/v0.0.1...v0.0.6
[0.0.1]: https://github.com/scanoss/licenses/releases/tag/v0.0.1
