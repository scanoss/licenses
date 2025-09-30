DROP TABLE IF EXISTS versions;
CREATE TABLE versions
(
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    version_name text    not null unique,
    semver       text default ''
);
insert into versions (id, version_name, semver) values(1,'1.0.0','v1.0.0');


