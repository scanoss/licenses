-- Test data only. The schema comes from testSchemaDDL in pkg/model/test_schema.go; do not add DDL here.
-- Serving `component_license`: a component's licence with no release attached. It
-- is the fallback the lookups fall to once no release has a licence.
--
-- pkg:cargo/serde has no component_version_license row, so it can only be answered
-- from here. pkg:npm/express has several, so its rows here must stay unused - they
-- are present to prove the fallback does not override per-release data.
INSERT INTO component_license (purl_id, license_id) VALUES ('3d9c7a14-2b6e-5f08-9c41-7ae0d5b12f63', 552);
INSERT INTO component_license (purl_id, license_id) VALUES ('3d9c7a14-2b6e-5f08-9c41-7ae0d5b12f63', 109);
INSERT INTO component_license (purl_id, license_id) VALUES ('eea6a5f0-ac07-5faa-86d6-42a1f88f80bd', 15);
