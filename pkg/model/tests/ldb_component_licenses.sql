-- Test data for ldb_component_licenses table
-- SPDX-License-Identifier: GPL-2.0-or-later

CREATE TABLE IF NOT EXISTS ldb_component_licenses (
    purl_md5 TEXT NOT NULL,
    source TEXT,
    license_id INTEGER NOT NULL,
    FOREIGN KEY (license_id) REFERENCES licenses(id)
);

-- Insert test data
INSERT INTO ldb_component_licenses (purl_md5, source, license_id) VALUES
-- Test case 1: Single license for a component (MIT)
('abc123def456789', 'github.com/example/repo', 5614),

-- Test case 2: Multiple licenses for same component (different sources)
('xyz789abc123456', 'github.com/multi/repo', 5614),     -- MIT
('xyz789abc123456', 'npmjs.com/multi-package', 552),    -- Apache 2.0

-- Test case 3: Component with GPL license
('gpl123component', 'gitlab.com/gpl/project', 2815),    -- GPL-2.0-only

-- Test case 4: Component with multiple licenses from same source
('dual456license', 'github.com/dual/licensed', 5614),   -- MIT
('dual456license', 'github.com/dual/licensed', 552),    -- Apache 2.0

-- Test case 5: Another single license component for boundary testing (BSD)
('test789boundary', 'bitbucket.org/boundary/test', 109), -- 3-Clause BSD License

-- Test case 6: Component with non-existent license (to test LEFT JOIN behavior)
('orphan456license', 'github.com/orphan/project', 99999);