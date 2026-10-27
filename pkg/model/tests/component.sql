-- Test data only. The schema comes from testSchemaDDL in pkg/model/test_schema.go; do not add DDL here.
-- Serving `component`: the identity the licence lookups key on, replacing the
-- purl string mining's purl_licenses carried on every row.
INSERT INTO component (purl, vendor, component, purl_id, source_purl_id) VALUES ('pkg:gem/rails', NULL, NULL, '4cdef47c-04e6-50cf-8acb-a8544fe5d367', NULL);
INSERT INTO component (purl, vendor, component, purl_id, source_purl_id) VALUES ('pkg:gitlab/gpl/project', NULL, NULL, '744c82ff-9c50-5fcd-aef7-8bd70fa70062', NULL);
INSERT INTO component (purl, vendor, component, purl_id, source_purl_id) VALUES ('pkg:maven/org.apache.commons/commons-lang3', NULL, NULL, '15baa372-3d76-5c24-b93b-f7862f8f59f6', NULL);
INSERT INTO component (purl, vendor, component, purl_id, source_purl_id) VALUES ('pkg:npm/express', NULL, NULL, 'eea6a5f0-ac07-5faa-86d6-42a1f88f80bd', NULL);
INSERT INTO component (purl, vendor, component, purl_id, source_purl_id) VALUES ('pkg:npm/lodash', NULL, NULL, '2c3f66b0-d870-5a65-b33f-103286472769', NULL);
INSERT INTO component (purl, vendor, component, purl_id, source_purl_id) VALUES ('pkg:pypi/numpy', NULL, NULL, '07cd3fce-8d79-5510-a76e-6c741a92c8ce', NULL);
INSERT INTO component (purl, vendor, component, purl_id, source_purl_id) VALUES ('pkg:pypi/requests', NULL, NULL, 'b2612508-69f0-5763-aca7-6c87ba8563a9', NULL);
-- pkg:cargo/serde has no row in component_version_license: it stands for the 8.9
-- million components whose only licence is the package-level one.
INSERT INTO component (purl, vendor, component, purl_id, source_purl_id) VALUES ('pkg:cargo/serde', NULL, NULL, '3d9c7a14-2b6e-5f08-9c41-7ae0d5b12f63', NULL);
