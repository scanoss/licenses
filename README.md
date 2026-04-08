# SCANOSS Platform 2.0 Licenses
Welcome to the licenses server for SCANOSS Platform 2.0

**Warning** Work In Progress **Warning**

## Repository Structure
This repository is made up of the following components:
* ?

## Configuration

Environmental variables are fed in this order:

dot-env --> env.json -->  Actual Environment Variable

These are the supported configuration arguments:

```
APP_NAME="SCANOSS License Server"
APP_PORT=50051
APP_MODE=dev
APP_DEBUG=false

DB_DRIVER=postgres
DB_HOST=localhost
DB_USER=scanoss
DB_PASSWD=
DB_SCHEMA=scanoss
DB_SSL_MODE=disable
DB_DSN=

LOOKUP_SOURCE_PRIORITY=0,31,32,33,3,5
```

### License lookup source priority

`LOOKUP_SOURCE_PRIORITY` is an **ordered** list of license detection source IDs. When resolving licenses for a component, the service walks the list from highest to lowest priority and **stops at the first source that returns data** — lower-priority sources are only consulted when the current one yields no rows. The order you write is the priority order.

The default value is `[0, 31, 32, 33, 34, 35, 3, 5]`, which corresponds to the following sources:

| ID | Source               | Scope     | Description                                                                                                              |
|----|----------------------|-----------|--------------------------------------------------------------------------------------------------------------------------|
| 0  | `component_declared` | component | Declared at repository — read from project metadata via URL discovery using the API. Reported as the **first** source.  |
| 31 | `license_file`       | component | Detected at attribution file — files in the root directory whose lowercased names match `notice`, `license`, `readme`, or `copying`. Reported as the **second** source. |
| 32 | `license_file`       | component | Detected at the `./license/` folder — all files inside the `license` subfolder of the root directory. Internal use.     |
| 33 | `metadata_file`      | component | Detected at metadata file — files inside `./meta-inf/` whose lowercased names match `notice`, `license`, `readme`, `copying`, or `manifest`. Reported as the **fourth** source. |
| 34 | `licenses_folder`    | component | Detected at the `./licenses/` folder — all files inside the `licenses` subfolder of the root directory. Internal use. |
| 35 | `component_declared` | component | Scraped from the repository/API using the internal scraping tool (on-demand, fills gaps). Reported as the **third** source. |
| 3  | `license_file`       | component | Legacy back-compat detection in attribution files (pending remediation). Reported as the **second** source. |
| 5  | `scancode`           | component | Detected on attribution file by ScanCode — component-level licenses extracted from attribution files via ScanCode. Reported as the **last** source. |

Can also be set in the JSON config file:

```json
{
  "Lookup": {
    "SourcePriority": [0, 31, 32, 33, 34, 35, 3, 5]
  }
}
```

Whitespace after commas is tolerated. An empty list causes the service to fail at startup.


## Docker Environment

The license server can be deployed as a Docker container.

Adjust configurations by updating an .env file in the root of this repository.


### How to build

You can build your own image of the SCANOSS License Server with the ```docker build``` command as follows.

```bash
make ghcr_build
```


### How to run

Run the SCANOSS License Server Docker image by specifying the environmental file to be used with the ```--env-file``` argument. 

You may also need to expose the ```APP_PORT``` on a given ```interface:port``` with the ```-p``` argument.

```bash
docker run -it -v "$(pwd)":"$(pwd)" -p 50051:50051 ghcr.io/scanoss/scanoss-licenses -json-config $(pwd)/config/app-config-docker-local-dev.json -debug
```

## Development

To run locally on your desktop, please use the following command:

```shell
go run cmd/server/main.go -json-config config/app-config-dev.json -debug
```

After changing a license version, please run the following command:
```shell
go mod tidy -compat=1.19
```
https://mholt.github.io/json-to-go/

## Bugs/Features
To request features or alert about bugs, please do so [here](https://github.com/scanoss/dependencies/issues).

## Changelog
Details of major changes to the library can be found in [CHANGELOG.md](CHANGELOG.md).
