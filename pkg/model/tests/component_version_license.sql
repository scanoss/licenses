-- Test data only. The schema comes from testSchemaDDL in pkg/model/test_schema.go; do not add DDL here.
-- Serving `component_version_license`, keyed by (purl_id, version). The mining
-- `date` column has no equivalent and is dropped; `source` keeps the mining
-- source_id so the priority resolution has something to exercise, even though
-- the production export currently writes 0 on every row.
INSERT INTO component_version_license (purl_id, version, license_id, source) VALUES ('eea6a5f0-ac07-5faa-86d6-42a1f88f80bd', '4.18.2', 5614, 1);
INSERT INTO component_version_license (purl_id, version, license_id, source) VALUES ('eea6a5f0-ac07-5faa-86d6-42a1f88f80bd', '4.18.2', 552, 2);
INSERT INTO component_version_license (purl_id, version, license_id, source) VALUES ('eea6a5f0-ac07-5faa-86d6-42a1f88f80bd', '4.17.1', 5614, 1);
INSERT INTO component_version_license (purl_id, version, license_id, source) VALUES ('2c3f66b0-d870-5a65-b33f-103286472769', '4.17.21', 5614, 1);
INSERT INTO component_version_license (purl_id, version, license_id, source) VALUES ('2c3f66b0-d870-5a65-b33f-103286472769', '4.17.21', 109, 2);
INSERT INTO component_version_license (purl_id, version, license_id, source) VALUES ('15baa372-3d76-5c24-b93b-f7862f8f59f6', '3.12.0', 552, 1);
INSERT INTO component_version_license (purl_id, version, license_id, source) VALUES ('15baa372-3d76-5c24-b93b-f7862f8f59f6', '3.12.0', 850, 3);
INSERT INTO component_version_license (purl_id, version, license_id, source) VALUES ('b2612508-69f0-5763-aca7-6c87ba8563a9', '2.28.1', 552, 1);
INSERT INTO component_version_license (purl_id, version, license_id, source) VALUES ('07cd3fce-8d79-5510-a76e-6c741a92c8ce', '1.24.0', 109, 1);
INSERT INTO component_version_license (purl_id, version, license_id, source) VALUES ('4cdef47c-04e6-50cf-8acb-a8544fe5d367', '7.0.4', 5614, 1);
INSERT INTO component_version_license (purl_id, version, license_id, source) VALUES ('744c82ff-9c50-5fcd-aef7-8bd70fa70062', '1.0.0', 2815, 0);
INSERT INTO component_version_license (purl_id, version, license_id, source) VALUES ('744c82ff-9c50-5fcd-aef7-8bd70fa70062', '1.0.0', 2815, 31);
INSERT INTO component_version_license (purl_id, version, license_id, source) VALUES ('744c82ff-9c50-5fcd-aef7-8bd70fa70062', '1.0.0', 2815, 6);
