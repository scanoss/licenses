DROP TABLE IF EXISTS all_urls;
CREATE TABLE all_urls
(
    package_hash text not null,
    vendor       text,
    component    text,
    version      text,
    date         text,
    url          text not null,
    url_hash     text not null,
    mine_id      integer,
    license      text,
    version_id   integer,
    license_id   integer,
    purl_name    text,
    is_mined     boolean,
    primary key (package_hash, url, url_hash)
);

INSERT INTO all_urls (package_hash, url_hash,vendor, component, version, date,  url,mine_id, purl_name, version_id, license_id) values ('c8b5647654826091fb65a97bec820eb9','c8b5647654826091fb65a97bec820eb9', 'pineappleea','pineapple-src','v1.0','2024-06-24','https://github.com/pineappleea/pineapple-src',5,'pineappleea/pineapple-src',1,83);
INSERT INTO all_urls (package_hash, vendor, component, version, date, url, url_hash, mine_id, license, purl_name, version_id, license_id) values ('abc123def456ghi789jkl012mno345pq', 'gpl', 'project', '1.0.0', '2025-09-30', 'https://gitlab.com/gpl/project', 'xyz789abc123def456ghi789jkl012mn', 39, 'GPL-2.0-only', 'gpl/project', 1, 2815);
