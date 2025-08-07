DROP TABLE IF EXISTS purl_licenses;
CREATE TABLE purl_licenses
(
    purl       text     not null,
    version    text     not null,
    date       text     not null,
    source_id  integer  not null,
    license_id integer  not null,
    unique (purl, version, source_id, license_id)
);

INSERT INTO purl_licenses (purl, version, date, source_id, license_id) VALUES
('pkg:npm/express', '4.18.2', '2023-01-01', 1, 5614),
('pkg:npm/express', '4.18.2', '2023-01-01', 2, 552),
('pkg:npm/express', '4.17.1', '2022-12-01', 1, 5614),
('pkg:npm/lodash', '4.17.21', '2023-01-02', 1, 5614),
('pkg:npm/lodash', '4.17.21', '2023-01-02', 2, 109),
('pkg:maven/org.apache.commons/commons-lang3', '3.12.0', '2023-01-03', 1, 552),
('pkg:maven/org.apache.commons/commons-lang3', '3.12.0', '2023-01-03', 3, 850),
('pkg:pypi/requests', '2.28.1', '2023-01-04', 1, 552),
('pkg:pypi/numpy', '1.24.0', '2023-01-05', 1, 109),
('pkg:gem/rails', '7.0.4', '2023-01-06', 1, 5614);