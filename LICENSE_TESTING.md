# License Service Testing

## Service Endpoint
```
curl -s -X POST http://168.119.136.95:40059/api/v2/licenses/purl
```

## Test Components Matrix

### Dual-Licensed Components
| Component | PURL | Expected License | Notes |
|-----------|------|------------------|-------|
| Qt Framework | `pkg:github/qt/qtbase` | GPL/Commercial | Major GUI framework |
| MySQL Connector | `pkg:pypi/mysql-connector-python` | GPL/Commercial | Database connector |
| FFmpeg | `pkg:github/ffmpeg/ffmpeg` | GPL/LGPL | Media processing |
| Blender | `pkg:github/blender/blender` | GPL with Apache parts | 3D modeling |
| OpenSSL | `pkg:github/openssl/openssl` | Apache 2.0/OpenSSL License | Cryptography |
| Redis | `pkg:github/redis/redis` | BSD/SSPL | In-memory database |

### Single License Reference Cases
| Component | PURL | Expected License | Notes |
|-----------|------|------------------|-------|
| React | `pkg:npm/react` | MIT | UI library |
| Angular Core | `pkg:npm/@angular/core` | MIT | Framework |
| SQLite | `pkg:github/sqlite/sqlite` | Public Domain | Database |
| PostgreSQL | `pkg:github/postgres/postgres` | PostgreSQL License | Database |

### Version Testing
| Component | PURL with Version | Notes |
|-----------|-------------------|-------|
| FFmpeg v4.4 | `pkg:github/ffmpeg/ffmpeg@4.4` | Specific version |
| React v18 | `pkg:npm/react@18.2.0` | Specific version |
| Qt Latest | `pkg:github/qt/qtbase@latest` | Latest tag |

## Test Results

### Dual-Licensed Components Results
| Component | PURL | Detected License | Version Found | Status | Notes |
|-----------|------|------------------|---------------|---------|-------|
| Qt Framework | `pkg:github/qt/qtbase` | ❌ No licenses | - | ⚠️ Issue | Expected GPL/Commercial dual license |
| MySQL Connector | `pkg:pypi/mysql-connector-python` | ❌ No licenses | - | ⚠️ Issue | Expected GPL/Commercial dual license |
| FFmpeg | `pkg:github/ffmpeg/ffmpeg` | ✅ GPL-2.0-only AND GPL-3.0-only AND LGPL-2.1-only AND LGPL-3.0-or-later | n7.2-dev | ✅ Success | Correctly detected dual/multi license |
| Blender | `pkg:github/blender/blender` | ⚠️ BSD-3-Clause only | v4.5.1 | ⚠️ Partial | Missing GPL parts, only detected BSD |
| OpenSSL | `pkg:github/openssl/openssl` | ✅ Apache-2.0 | openssl-3.5.2 | ✅ Success | Correctly detected (OpenSSL License deprecated) |
| Redis | `pkg:github/redis/redis` | ⚠️ BSD-3-Clause only | twitter-20100825 | ⚠️ Partial | Missing SSPL, old version detected |

### Single License Results
| Component | PURL | Detected License | Version Found | Status | Notes |
|-----------|------|------------------|---------------|---------|-------|
| React | `pkg:npm/react` | ✅ MIT | 19.2.0-canary-fa3feba6-20250623 | ✅ Success | Correctly detected |
| Angular Core | `pkg:npm/@angular/core` | ❌ No licenses | - | ⚠️ Issue | Expected MIT license |
| SQLite | `pkg:github/sqlite/sqlite` | ❌ No licenses | - | ⚠️ Issue | Expected Public Domain |
| PostgreSQL | `pkg:github/postgres/postgres` | ❌ No licenses | - | ⚠️ Issue | Expected PostgreSQL License |

### Version Testing Results
| Component | PURL with Version | Detected License | Status | Notes |
|-----------|-------------------|------------------|---------|-------|
| FFmpeg v4.4 | `pkg:github/ffmpeg/ffmpeg@4.4` | ❌ No licenses | ❌ Failed | Version-specific lookup failed |
| React v18.2.0 | `pkg:npm/react@18.2.0` | ✅ MIT | ✅ Success | Version-specific lookup worked |
| Qt Latest | `pkg:github/qt/qtbase@latest` | ❌ No licenses | ❌ Failed | Latest tag lookup failed |

## Issues Found

### Major Issues
1. **Missing License Data**: Several major components return empty license arrays:
    - Qt Framework (dual-licensed)
    - MySQL Connector (dual-licensed)
    - Angular Core (MIT)
    - SQLite (Public Domain)
    - PostgreSQL (PostgreSQL License)

2. **Version Handling**:
    - GitHub components with specific versions often fail to return license data
    - NPM components with versions work better (React@18.2.0 succeeded)

3. **Incomplete Dual License Detection**:
    - Blender only shows BSD-3-Clause, missing GPL portions
    - Redis only shows BSD-3-Clause, missing newer SSPL license

### Successful Cases
1. **FFmpeg**: Excellent multi-license detection (GPL-2.0, GPL-3.0, LGPL-2.1, LGPL-3.0)
2. **React**: Consistent MIT detection across versions
3. **OpenSSL**: Correct Apache-2.0 detection

### Service Behavior Analysis
- ✅ Service responds successfully (HTTP 200)
- ✅ JSON format is consistent
- ⚠️ License detection coverage is inconsistent
- ⚠️ Version handling varies by package type (NPM > GitHub)
- ⚠️ Some major open-source projects not in database